package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/uuid"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
)

const (
	xunleiAPI     = "https://x-api-pan.xunlei.com/drive/v1"
	xunleiUserAPI = "https://xluser-ssl.xunlei.com/v1"
)

const (
	xunleiDefaultUA = "AndroidDownloadManager/13 (Linux; U; Android 13; M2004J7AC Build/SP1A.210812.016)"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "xunlei",
		Name:        "迅雷网盘",
		AuthType:    cloudsync.AuthRefreshToken,
		Description: "refresh_token 登录，与迅雷下载生态打通",
	}, enum.DownloadTaskTypeXunlei.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewXunleiProvider(cookie)
	})
}

type xunleiFile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Size           string `json:"size"`
	WebContentLink string `json:"web_content_link"`
	Kind           string `json:"kind"`
	Trashed        bool   `json:"trashed"`
	ParentID       string `json:"parent_id"`
}

type xunleiListResp struct {
	Files         []xunleiFile `json:"files"`
	NextPageToken string       `json:"next_page_token"`
}

type xunleiTokenResp struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
}

type xunleiUserInfo struct {
	Name  string `json:"name"`
	Sub   string `json:"sub"`
	Space string `json:"space"`
}

// XunleiProvider 迅雷网盘提供者
type XunleiProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	tokenType    string
	deviceID     string
	mu           sync.RWMutex
	client       *http.Client
	credSync     credentialSyncHelper
}

// NewXunleiProvider 创建迅雷网盘提供者
func NewXunleiProvider(credential string) *XunleiProvider {
	cred := internal.ParseOAuthCredential(credential, cloudsync.XunleiClientID, cloudsync.XunleiClientSecret)
	return &XunleiProvider{
		refreshToken: cred.RefreshToken,
		clientID:     cred.ClientID,
		clientSecret: cred.ClientSecret,
		deviceID:     uuid.New().String(),
		client:       internal.DefaultHTTPClient(),
	}
}

func (x *XunleiProvider) Name() string {
	return "xunlei"
}

func (x *XunleiProvider) SetCredentialUpdateCallback(fn func(string)) {
	x.credSync.SetCredentialUpdateCallback(fn)
}

func (x *XunleiProvider) ExportCredential() string {
	x.mu.RLock()
	defer x.mu.RUnlock()
	return internal.FormatOAuthCredential(x.refreshToken, x.clientID, x.clientSecret, cloudsync.XunleiClientID, cloudsync.XunleiClientSecret)
}

func (x *XunleiProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	var userInfo xunleiUserInfo
	if err := x.get(xunleiUserAPI+"/user/me", nil, &userInfo); err != nil {
		return nil, err
	}

	nickname := userInfo.Name
	if nickname == "" {
		nickname = "迅雷用户"
	}

	return &cloudsync.CloudUserInfo{
		Nickname: nickname,
	}, nil
}

func (x *XunleiProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	var pageToken string
	for i := 0; i < page; i++ {
		params := map[string]string{
			"parent_id": pdirFid,
			"limit":     fmt.Sprintf("%d", size),
			"filters":   `{"trashed":{"eq":false}}`,
		}
		if pageToken != "" {
			params["page_token"] = pageToken
		}

		var resp xunleiListResp
		if err := x.get(xunleiAPI+"/files", params, &resp); err != nil {
			return nil, 0, err
		}

		if i == page-1 {
			files := make([]cloudsync.CloudFile, 0, len(resp.Files))
			for _, f := range resp.Files {
				if f.Trashed {
					continue
				}
				files = append(files, cloudsync.CloudFile{
					Fid:      f.ID,
					FileName: f.Name,
					Size:     parseSize(f.Size),
					IsDir:    f.Kind == "drive#folder",
				})
			}
			total := (page-1)*size + len(resp.Files)
			if resp.NextPageToken != "" {
				total = page*size + 1
			}
			return files, total, nil
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return nil, 0, nil
		}
	}

	return nil, 0, nil
}

func (x *XunleiProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := x.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	var file xunleiFile
	if err := x.get(xunleiAPI+"/files/"+fid, nil, &file); err != nil {
		return nil, err
	}

	if file.WebContentLink == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	return &cloudsync.CloudDownloadLink{
		DownloadURL: file.WebContentLink,
		Headers: map[string]string{
			"User-Agent": xunleiDefaultUA,
		},
	}, nil
}

func (x *XunleiProvider) ensureAccessToken() error {
	x.mu.RLock()
	if x.accessToken != "" {
		x.mu.RUnlock()
		return nil
	}
	x.mu.RUnlock()
	return x.refreshAccessToken()
}

func (x *XunleiProvider) refreshAccessToken() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.accessToken != "" {
		return nil
	}

	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": x.refreshToken,
		"client_id":     x.clientID,
		"client_secret": x.clientSecret,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", xunleiUserAPI+"/auth/token", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取token响应失败: %w", err)
	}

	var tokenResp xunleiTokenResp
	if err := json.Unmarshal(respBytes, &tokenResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("token刷新失败: access_token为空")
	}

	x.accessToken = tokenResp.AccessToken
	x.tokenType = tokenResp.TokenType
	if tokenResp.RefreshToken != "" {
		x.refreshToken = tokenResp.RefreshToken
	}
	x.credSync.notifyCredential(x.refreshToken, x.clientID, x.clientSecret, cloudsync.XunleiClientID, cloudsync.XunleiClientSecret)
	return nil
}

func (x *XunleiProvider) invalidateToken() {
	x.mu.Lock()
	x.accessToken = ""
	x.mu.Unlock()
}

func (x *XunleiProvider) get(url string, params map[string]string, result interface{}) error {
	maxRetries := 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		x.mu.RLock()
		token := x.accessToken
		tokenType := x.tokenType
		x.mu.RUnlock()

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Authorization", tokenType+" "+token)
		req.Header.Set("User-Agent", xunleiDefaultUA)
		req.Header.Set("Accept", "application/json;charset=UTF-8")
		req.Header.Set("x-device-id", x.deviceID)
		req.Header.Set("x-client-id", x.clientID)
		req.Header.Set("x-client-version", "1.10.0.2633")

		resp, err := x.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		errCode := internal.JSONIntField(respBytes, "error_code")
		if errCode != 0 {
			if errCode == 4122 || errCode == 4121 || errCode == 10 || errCode == 16 {
				x.invalidateToken()
				if err := x.refreshAccessToken(); err != nil {
					return fmt.Errorf("token刷新失败: %w", err)
				}
				continue
			}
			errMsg := internal.JSONStringField(respBytes, "error_description")
			if errMsg == "" {
				errMsg = internal.JSONStringField(respBytes, "message")
			}
			return fmt.Errorf("迅雷API错误: error_code=%d, %s", errCode, errMsg)
		}

		if result != nil {
			if err := json.Unmarshal(respBytes, result); err != nil {
				return fmt.Errorf("解析响应失败: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("请求重试%d次后仍失败", maxRetries)
}

func parseSize(s string) int64 {
	var size int64
	fmt.Sscanf(s, "%d", &size)
	return size
}
