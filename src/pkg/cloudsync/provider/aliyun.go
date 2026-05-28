package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
)

const (
	aliyunAPI = "https://openapi.alipan.com"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "aliyun",
		Name:        "阿里云盘",
		AuthType:    cloudsync.AuthRefreshToken,
		Description: "refresh_token 登录；需在 config.toml [cloud] 配置 aliyun_client_id/secret，或凭据格式 refresh_token|client_id|client_secret",
	}, enum.DownloadTaskTypeAliyun.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewAliyunProvider(cookie)
	})
}

// aliyunErrResp 错误响应
type aliyunErrResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// aliyunFile 阿里云盘文件
type aliyunFile struct {
	DriveId       string `json:"drive_id"`
	FileId        string `json:"file_id"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Type          string `json:"type"`
	FileExtension string `json:"file_extension"`
}

// aliyunListResp 文件列表响应
type aliyunListResp struct {
	Items      []aliyunFile `json:"items"`
	NextMarker string       `json:"next_marker"`
}

// aliyunDownloadResp 下载链接响应
type aliyunDownloadResp struct {
	URL string `json:"url"`
}

// aliyunDriveInfo 云盘信息
type aliyunDriveInfo struct {
	TotalSize int64 `json:"total_size"`
	UsedSize  int64 `json:"used_size"`
}

// AliyunProvider 阿里云盘提供者（Open API）
type AliyunProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	driveID      string
	mu           sync.RWMutex
	client       *http.Client
	credSync     credentialSyncHelper
}

// NewAliyunProvider 创建阿里云盘提供者
func NewAliyunProvider(credential string) *AliyunProvider {
	cred := internal.ParseOAuthCredential(credential, cloudsync.AliyunClientID, cloudsync.AliyunClientSecret)
	return &AliyunProvider{
		refreshToken: cred.RefreshToken,
		clientID:     cred.ClientID,
		clientSecret: cred.ClientSecret,
		client:       internal.DefaultHTTPClient(),
	}
}

func (a *AliyunProvider) Name() string {
	return "aliyun"
}

func (a *AliyunProvider) SetCredentialUpdateCallback(fn func(string)) {
	a.credSync.SetCredentialUpdateCallback(fn)
}

func (a *AliyunProvider) ExportCredential() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return internal.FormatOAuthCredential(a.refreshToken, a.clientID, a.clientSecret, cloudsync.AliyunClientID, cloudsync.AliyunClientSecret)
}

func (a *AliyunProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	var driveInfo aliyunDriveInfo
	if err := a.post("/adrive/v1.0/user/getDriveInfo", nil, &driveInfo); err != nil {
		return nil, err
	}

	var userInfo struct {
		NickName string `json:"nick_name"`
	}
	a.post("/adrive/v1.0/user/get", nil, &userInfo)

	nickname := userInfo.NickName
	if nickname == "" {
		nickname = "阿里云盘用户"
	}

	return &cloudsync.CloudUserInfo{
		Nickname:  nickname,
		TotalSize: driveInfo.TotalSize,
		UsedSize:  driveInfo.UsedSize,
	}, nil
}

func (a *AliyunProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}

	if pdirFid == "" {
		pdirFid = "root"
	}

	var marker string
	for i := 0; i < page; i++ {
		body := map[string]interface{}{
			"drive_id":        a.driveID,
			"parent_file_id":  pdirFid,
			"limit":           size,
			"fields":          "*",
			"order_by":        "updated_at",
			"order_direction": "DESC",
		}
		if marker != "" {
			body["marker"] = marker
		}

		var resp aliyunListResp
		if err := a.post("/adrive/v1.0/openFile/list", body, &resp); err != nil {
			return nil, 0, err
		}

		if i == page-1 {
			files := make([]cloudsync.CloudFile, 0, len(resp.Items))
			for _, f := range resp.Items {
				files = append(files, cloudsync.CloudFile{
					Fid:      f.FileId,
					FileName: f.Name,
					Size:     f.Size,
					IsDir:    f.Type == "folder",
				})
			}
			total := (page-1)*size + len(resp.Items)
			if resp.NextMarker != "" {
				total = page*size + 1
			}
			return files, total, nil
		}

		marker = resp.NextMarker
		if marker == "" {
			return nil, 0, nil
		}
	}

	return nil, 0, nil
}

func (a *AliyunProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := a.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	body := map[string]interface{}{
		"drive_id":   a.driveID,
		"file_id":    fid,
		"expire_sec": 14400,
	}

	var resp aliyunDownloadResp
	if err := a.post("/adrive/v1.0/openFile/getDownloadUrl", body, &resp); err != nil {
		return nil, err
	}

	if resp.URL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}

	expires := time.Now().Add(4 * time.Hour)
	return &cloudsync.CloudDownloadLink{
		DownloadURL: resp.URL,
		ExpiresAt:   &expires,
	}, nil
}

func (a *AliyunProvider) ensureAccessToken() error {
	a.mu.RLock()
	if a.accessToken != "" {
		a.mu.RUnlock()
		return nil
	}
	a.mu.RUnlock()
	return a.refreshAccessToken()
}

func (a *AliyunProvider) refreshAccessToken() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" {
		return nil
	}

	if a.clientID == "" || a.clientSecret == "" {
		return fmt.Errorf("阿里云盘需要 client_id 和 client_secret，请在 config.toml [cloud] 配置 aliyun_client_id/secret，或使用凭据格式 refresh_token|client_id|client_secret")
	}

	body := map[string]string{
		"refresh_token": a.refreshToken,
		"grant_type":    "refresh_token",
		"client_id":     a.clientID,
		"client_secret": a.clientSecret,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", aliyunAPI+"/oauth/access_token", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取token响应失败: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBytes, &tokenResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("token刷新失败: access_token为空")
	}

	a.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		a.refreshToken = tokenResp.RefreshToken
	}
	a.credSync.notifyCredential(a.refreshToken, a.clientID, a.clientSecret, cloudsync.AliyunClientID, cloudsync.AliyunClientSecret)
	a.getDriveID()
	return nil
}

func (a *AliyunProvider) getDriveID() {
	var resp struct {
		DefaultDriveID string `json:"default_drive_id"`
	}
	if err := a.post("/adrive/v1.0/user/getDriveInfo", nil, &resp); err == nil {
		a.driveID = resp.DefaultDriveID
	}
}

func (a *AliyunProvider) invalidateToken() {
	a.mu.Lock()
	a.accessToken = ""
	a.mu.Unlock()
}

func (a *AliyunProvider) post(path string, body interface{}, result interface{}) error {
	maxRetries := 2

	for attempt := 0; attempt < maxRetries; attempt++ {
		a.mu.RLock()
		token := a.accessToken
		a.mu.RUnlock()

		var reqBody io.Reader
		if body != nil {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("序列化请求体失败: %w", err)
			}
			reqBody = bytes.NewReader(bodyBytes)
		} else {
			reqBody = bytes.NewReader([]byte("{}"))
		}

		req, err := http.NewRequest("POST", aliyunAPI+path, reqBody)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		var errResp aliyunErrResp
		if err := json.Unmarshal(respBytes, &errResp); err == nil && errResp.Code != "" {
			if errResp.Code == "AccessTokenInvalid" || errResp.Code == "AccessTokenExpired" {
				a.invalidateToken()
				if err := a.refreshAccessToken(); err != nil {
					return fmt.Errorf("token刷新失败: %w", err)
				}
				continue
			}
			return fmt.Errorf("阿里云盘API错误: %s - %s", errResp.Code, errResp.Message)
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
