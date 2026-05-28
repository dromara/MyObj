package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
)

const pan123APIBase = "https://open-api.123pan.com"

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "123pan",
		Name:        "123云盘",
		AuthType:    cloudsync.AuthRefreshToken,
		Description: "123 开放平台 client_id + client_secret 登录；可在 config.toml [cloud] 配置 pan123_client_id/secret",
		CredentialFields: []cloudsync.CredentialField{
			{Key: "client_id", Label: "应用 ID (Client ID)", Required: true, Help: "也可在 config.toml [cloud] 配置 pan123_client_id"},
			{Key: "client_secret", Label: "应用密钥 (Client Secret)", Required: true, Secret: true, Help: "也可在 config.toml [cloud] 配置 pan123_client_secret"},
		},
	}, enum.DownloadTaskTypePan123.Value(), func(credential string) cloudsync.CloudProvider {
		return NewPan123Provider(credential)
	})
}

type pan123File struct {
	FileID   int64  `json:"fileId"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     int    `json:"type"`
	Trashed  int    `json:"trashed"`
}

type pan123ListResp struct {
	Code int `json:"code"`
	Data struct {
		FileList   []pan123File `json:"fileList"`
		LastFileID int64        `json:"lastFileId"`
	} `json:"data"`
	Message string `json:"message"`
}

type Pan123Provider struct {
	clientID     string
	clientSecret string
	accessToken  string
	expiresAt    time.Time
	mu           sync.RWMutex
	client       *http.Client
}

func NewPan123Provider(credential string) *Pan123Provider {
	parts := strings.SplitN(strings.TrimSpace(credential), "|", 2)
	p := &Pan123Provider{
		clientID:     cloudsync.Pan123ClientID,
		clientSecret: cloudsync.Pan123ClientSecret,
		client:       internal.DefaultHTTPClient(),
	}
	if len(parts) >= 1 && parts[0] != "" {
		p.clientID = parts[0]
	}
	if len(parts) >= 2 && parts[1] != "" {
		p.clientSecret = parts[1]
	}
	return p
}

func (p *Pan123Provider) Name() string { return "123pan" }

func (p *Pan123Provider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}
	return &cloudsync.CloudUserInfo{Nickname: "123云盘用户"}, nil
}

func (p *Pan123Provider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := p.ensureToken(); err != nil {
		return nil, 0, err
	}
	if pdirFid == "" {
		pdirFid = "0"
	}
	parentID, _ := strconv.ParseInt(pdirFid, 10, 64)

	all := make([]pan123File, 0)
	cursor := int64(0)
	for {
		q := url.Values{
			"parentFileId": {strconv.FormatInt(parentID, 10)},
			"limit":        {"100"},
			"lastFileId":   {strconv.FormatInt(cursor, 10)},
		}
		var resp pan123ListResp
		if err := p.apiGet("/api/v2/file/list", q, &resp); err != nil {
			return nil, 0, err
		}
		for _, f := range resp.Data.FileList {
			if f.Trashed != 0 {
				continue
			}
			all = append(all, f)
		}
		if resp.Data.LastFileID == -1 || resp.Data.LastFileID == 0 {
			break
		}
		cursor = resp.Data.LastFileID
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}
	start := (page - 1) * size
	if start >= len(all) {
		return nil, len(all), nil
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	slice := all[start:end]

	files := make([]cloudsync.CloudFile, 0, len(slice))
	for _, f := range slice {
		files = append(files, cloudsync.CloudFile{
			Fid:      strconv.FormatInt(f.FileID, 10),
			FileName: f.Filename,
			Size:     f.Size,
			IsDir:    f.Type == 1,
		})
	}
	return files, len(all), nil
}

func (p *Pan123Provider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}
	q := url.Values{"fileID": {fid}}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := p.apiGet("/api/v1/direct-link/url", q, &resp); err != nil {
		return nil, err
	}
	if resp.Data.URL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}
	return &cloudsync.CloudDownloadLink{DownloadURL: resp.Data.URL}, nil
}

func (p *Pan123Provider) ensureToken() error {
	p.mu.RLock()
	if p.accessToken != "" && time.Now().Before(p.expiresAt) {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()
	return p.refreshToken()
}

func (p *Pan123Provider) refreshToken() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.clientID == "" || p.clientSecret == "" {
		return fmt.Errorf("123云盘需要 client_id 和 client_secret")
	}
	body := map[string]string{
		"clientID":     p.clientID,
		"clientSecret": p.clientSecret,
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, pan123APIBase+"/api/v1/access_token", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Platform", "open_platform")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Code int `json:"code"`
		Data struct {
			AccessToken string `json:"accessToken"`
			ExpiredAt   string `json:"expiredAt"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed.Code != 0 || parsed.Data.AccessToken == "" {
		return fmt.Errorf("123云盘 token 错误: %s", parsed.Message)
	}
	p.accessToken = parsed.Data.AccessToken
	p.expiresAt = time.Now().Add(50 * time.Minute)
	return nil
}

func (p *Pan123Provider) apiGet(path string, q url.Values, result interface{}) error {
	reqURL := pan123APIBase + path
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	p.mu.RLock()
	token := p.accessToken
	p.mu.RUnlock()
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Platform", "open_platform")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("123云盘 API 错误 %d: %s", resp.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, result)
}
