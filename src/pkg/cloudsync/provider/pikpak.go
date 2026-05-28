package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/internal"
	"myobj/src/pkg/enum"
)

const (
	pikpakAuthAPI  = "https://user.mypikpak.net/v1/auth/token"
	pikpakDriveAPI = "https://api-drive.mypikpak.net/drive/v1"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "pikpak",
		Name:        "PikPak",
		AuthType:    cloudsync.AuthRefreshToken,
		Description: "PikPak refresh_token 登录；可在 config.toml [cloud] 配置 pikpak_client_id/secret",
	}, enum.DownloadTaskTypePikPak.Value(), func(credential string) cloudsync.CloudProvider {
		return NewPikPakProvider(credential)
	})
}

type pikpakFile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Size           string `json:"size"`
	Kind           string `json:"kind"`
	WebContentLink string `json:"web_content_link"`
}

type pikpakListResp struct {
	Files         []pikpakFile `json:"files"`
	NextPageToken string       `json:"next_page_token"`
}

type PikPakProvider struct {
	refreshToken string
	clientID     string
	clientSecret string
	accessToken  string
	mu           sync.RWMutex
	client       *http.Client
	credSync     credentialSyncHelper
}

func NewPikPakProvider(credential string) *PikPakProvider {
	cred := internal.ParseOAuthCredential(credential, cloudsync.PikPakClientID, cloudsync.PikPakClientSecret)
	return &PikPakProvider{
		refreshToken: cred.RefreshToken,
		clientID:     cred.ClientID,
		clientSecret: cred.ClientSecret,
		client:       internal.DefaultHTTPClient(),
	}
}

func (p *PikPakProvider) SetCredentialUpdateCallback(fn func(string)) {
	p.credSync.SetCredentialUpdateCallback(fn)
}

func (p *PikPakProvider) ExportCredential() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return internal.FormatOAuthCredential(p.refreshToken, p.clientID, p.clientSecret, cloudsync.PikPakClientID, cloudsync.PikPakClientSecret)
}

func (p *PikPakProvider) Name() string { return "pikpak" }

func (p *PikPakProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, err
	}
	return &cloudsync.CloudUserInfo{Nickname: "PikPak用户"}, nil
}

func (p *PikPakProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}

	pageToken := ""
	for i := 0; i < page; i++ {
		q := url.Values{
			"parent_id": {pdirFid},
			"limit":     {fmt.Sprintf("%d", size)},
			"filters":   {`{"trashed":{"eq":false}}`},
		}
		if pageToken != "" {
			q.Set("page_token", pageToken)
		}
		var resp pikpakListResp
		if err := p.get(pikpakDriveAPI+"/files?"+q.Encode(), &resp); err != nil {
			return nil, 0, err
		}
		if i == page-1 {
			files := make([]cloudsync.CloudFile, 0, len(resp.Files))
			for _, f := range resp.Files {
				var sz int64
				fmt.Sscanf(f.Size, "%d", &sz)
				files = append(files, cloudsync.CloudFile{
					Fid:      f.ID,
					FileName: f.Name,
					Size:     sz,
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

func (p *PikPakProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	if err := p.ensureAccessToken(); err != nil {
		return nil, err
	}
	var file pikpakFile
	if err := p.get(pikpakDriveAPI+"/files/"+url.PathEscape(fid), &file); err != nil {
		return nil, err
	}
	if file.WebContentLink == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}
	var sz int64
	fmt.Sscanf(file.Size, "%d", &sz)
	return &cloudsync.CloudDownloadLink{
		DownloadURL: file.WebContentLink,
		FileName:    file.Name,
		Size:        sz,
	}, nil
}

func (p *PikPakProvider) ensureAccessToken() error {
	p.mu.RLock()
	if p.accessToken != "" {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()
	return p.refreshAccessToken()
}

func (p *PikPakProvider) refreshAccessToken() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.accessToken != "" {
		return nil
	}
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": p.refreshToken,
		"client_id":     p.clientID,
		"client_secret": p.clientSecret,
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, pikpakAuthAPI, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed.AccessToken == "" {
		return fmt.Errorf("PikPak token 刷新失败")
	}
	p.accessToken = parsed.AccessToken
	if parsed.RefreshToken != "" {
		p.refreshToken = parsed.RefreshToken
	}
	p.credSync.notifyCredential(p.refreshToken, p.clientID, p.clientSecret, cloudsync.PikPakClientID, cloudsync.PikPakClientSecret)
	return nil
}

func (p *PikPakProvider) get(reqURL string, result interface{}) error {
	for attempt := 0; attempt < 2; attempt++ {
		if err := p.ensureAccessToken(); err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return err
		}
		p.mu.RLock()
		token := p.accessToken
		p.mu.RUnlock()
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		resp, err := p.client.Do(req)
		if err != nil {
			return err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			p.mu.Lock()
			p.accessToken = ""
			p.mu.Unlock()
			if err := p.refreshAccessToken(); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("PikPak API 错误 %d: %s", resp.StatusCode, string(raw))
		}
		return json.Unmarshal(raw, result)
	}
	return fmt.Errorf("PikPak 请求失败")
}
