package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	pan115APIBase = "https://proapi.115.com"
	pan115AuthURL = "https://passportapi.115.com/open/refreshToken"
	pan115UserAPI = pan115APIBase + "/open/user/info"
	pan115FilesAPI = pan115APIBase + "/open/ufile/files"
	pan115DownURL = pan115APIBase + "/open/ufile/downurl"
	pan115DefaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID: "115", Name: "115网盘", AuthType: cloudsync.AuthRefreshToken,
		Description: "115 开放平台 refresh_token 登录；可在 config.toml [cloud] 配置 pan115_client_id/secret",
	}, enum.DownloadTaskTypePan115.Value(), func(credential string) cloudsync.CloudProvider {
		return NewPan115Provider(credential)
	})
}

type pan115File struct {
	Fid string `json:"fid"`
	Fn  string `json:"fn"`
	Fs  int64  `json:"fs"`
	Fc  string `json:"fc"` // 0=文件夹 1=文件
	Pc  string `json:"pc"` // pickcode，下载时使用
}

type pan115FilesResp struct {
	State   bool         `json:"state"`
	Code    int64        `json:"code"`
	Message string       `json:"message"`
	Data    []pan115File `json:"data"`
	Count   int64        `json:"count"`
}

type pan115UserResp struct {
	UserName    string `json:"user_name"`
	RtSpaceInfo struct {
		AllTotal struct {
			Size string `json:"size"`
		} `json:"all_total"`
		AllUse struct {
			Size string `json:"size"`
		} `json:"all_use"`
	} `json:"rt_space_info"`
}

type pan115DownItem struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	URL      struct {
		URL string `json:"url"`
	} `json:"url"`
}

// Pan115Provider 115 网盘提供者（开放平台 API）
type Pan115Provider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	mu           sync.RWMutex
	client       *http.Client
	credSync     credentialSyncHelper
}

// NewPan115Provider 创建 115 网盘提供者
// credential 格式: refresh_token 或 refresh_token|client_id|client_secret
func NewPan115Provider(credential string) *Pan115Provider {
	cred := internal.ParseOAuthCredential(credential, cloudsync.Pan115ClientID, cloudsync.Pan115ClientSecret)
	return &Pan115Provider{
		refreshToken: cred.RefreshToken,
		clientID:     cred.ClientID,
		clientSecret: cred.ClientSecret,
		client:       internal.DefaultHTTPClient(),
	}
}

func (p *Pan115Provider) SetCredentialUpdateCallback(fn func(string)) {
	p.credSync.SetCredentialUpdateCallback(fn)
}

func (p *Pan115Provider) ExportCredential() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return internal.FormatOAuthCredential(p.refreshToken, p.clientID, p.clientSecret, cloudsync.Pan115ClientID, cloudsync.Pan115ClientSecret)
}

func (p *Pan115Provider) Name() string {
	return "115"
}

func (p *Pan115Provider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}
	var user pan115UserResp
	if err := p.apiGet(pan115UserAPI, nil, &user, true); err != nil {
		return nil, err
	}
	nickname := user.UserName
	if nickname == "" {
		nickname = "115用户"
	}
	total, _ := strconv.ParseInt(user.RtSpaceInfo.AllTotal.Size, 10, 64)
	used, _ := strconv.ParseInt(user.RtSpaceInfo.AllUse.Size, 10, 64)
	return &cloudsync.CloudUserInfo{
		Nickname:  nickname,
		TotalSize: total,
		UsedSize:  used,
	}, nil
}

func (p *Pan115Provider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, 0, fmt.Errorf("刷新token失败: %w", err)
	}
	if pdirFid == "" {
		pdirFid = "0"
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}
	offset := (page - 1) * size

	var resp pan115FilesResp
	params := url.Values{
		"cid":      {pdirFid},
		"limit":    {strconv.Itoa(size)},
		"offset":   {strconv.Itoa(offset)},
		"show_dir": {"1"},
	}
	if err := p.apiGet(pan115FilesAPI, params, &resp, false); err != nil {
		return nil, 0, err
	}

	files := make([]cloudsync.CloudFile, 0, len(resp.Data))
	for _, f := range resp.Data {
		fid := f.Fid
		if f.Fc == "1" && f.Pc != "" {
			fid = f.Pc // 文件使用 pickcode 作为下载标识
		}
		files = append(files, cloudsync.CloudFile{
			Fid:      fid,
			FileName: f.Fn,
			Size:     f.Fs,
			IsDir:    f.Fc == "0",
		})
	}

	total := int(resp.Count)
	if total == 0 {
		total = offset + len(files)
		if len(files) == size {
			total = offset + size + 1
		}
	}
	return files, total, nil
}

func (p *Pan115Provider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}
	pickCode := fid
	if idx := strings.Index(fid, "|"); idx >= 0 {
		pickCode = fid[idx+1:]
	}

	form := url.Values{"pick_code": {pickCode}}
	var resp map[string]pan115DownItem
	if err := p.apiPostForm(pan115DownURL, form, &resp); err != nil {
		return nil, err
	}
	for _, item := range resp {
		if item.URL.URL == "" {
			continue
		}
		return &cloudsync.CloudDownloadLink{
			DownloadURL: item.URL.URL,
			FileName:    item.FileName,
			Size:        item.FileSize,
			Headers: map[string]string{
				"User-Agent": pan115DefaultUA,
			},
		}, nil
	}
	return nil, fmt.Errorf("未获取到下载链接")
}

func (p *Pan115Provider) ensureAccessToken() error {
	p.mu.RLock()
	if p.accessToken != "" {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()
	return p.refreshAccessToken()
}

func (p *Pan115Provider) refreshAccessToken() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.accessToken != "" {
		return nil
	}
	if p.refreshToken == "" {
		return fmt.Errorf("refresh_token 不能为空")
	}

	form := url.Values{"refresh_token": {p.refreshToken}}
	req, err := http.NewRequest(http.MethodPost, pan115AuthURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var authResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("解析token响应失败: %w", err)
	}
	if authResp.Code != 0 || authResp.Data.AccessToken == "" {
		msg := authResp.Message
		if msg == "" {
			msg = "token刷新失败"
		}
		return fmt.Errorf(msg)
	}
	p.accessToken = authResp.Data.AccessToken
	if authResp.Data.RefreshToken != "" {
		p.refreshToken = authResp.Data.RefreshToken
	}
	p.credSync.notifyCredential(p.refreshToken, p.clientID, p.clientSecret, cloudsync.Pan115ClientID, cloudsync.Pan115ClientSecret)
	return nil
}

func (p *Pan115Provider) invalidateToken() {
	p.mu.Lock()
	p.accessToken = ""
	p.mu.Unlock()
}

func (p *Pan115Provider) apiGet(api string, params url.Values, result interface{}, extractData bool) error {
	return p.doAuthRequest(http.MethodGet, api, params, nil, result, extractData)
}

func (p *Pan115Provider) apiPostForm(api string, form url.Values, result interface{}) error {
	return p.doAuthRequest(http.MethodPost, api, nil, form, result, true)
}

func (p *Pan115Provider) doAuthRequest(method, api string, query, form url.Values, result interface{}, extractData bool) error {
	for attempt := 0; attempt < 2; attempt++ {
		if err := p.ensureAccessToken(); err != nil {
			return err
		}

		reqURL := api
		if len(query) > 0 {
			reqURL += "?" + query.Encode()
		}

		var body io.Reader
		if form != nil {
			body = strings.NewReader(form.Encode())
		}
		req, err := http.NewRequest(method, reqURL, body)
		if err != nil {
			return err
		}
		p.mu.RLock()
		token := p.accessToken
		p.mu.RUnlock()
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", pan115DefaultUA)
		if form != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		resp, err := p.client.Do(req)
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		var envelope struct {
			State   bool            `json:"state"`
			Code    int64           `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
		if !envelope.State {
			if attempt == 0 && (envelope.Code == 99 || envelope.Code == 40140116) {
				p.invalidateToken()
				if err := p.refreshAccessToken(); err != nil {
					return err
				}
				continue
			}
			msg := envelope.Message
			if msg == "" {
				msg = fmt.Sprintf("115 API 错误 code=%d", envelope.Code)
			}
			return fmt.Errorf(msg)
		}

		if result == nil {
			return nil
		}
		if extractData && len(envelope.Data) > 0 {
			return json.Unmarshal(envelope.Data, result)
		}
		return json.Unmarshal(respBody, result)
	}
	return fmt.Errorf("115 API 请求重试后仍失败")
}
