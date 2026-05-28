package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/oauth"
	"myobj/src/pkg/enum"
)

const dropboxAPI = "https://api.dropboxapi.com/2"

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "dropbox",
		Name:        "Dropbox",
		AuthType:    cloudsync.AuthOAuth2,
		Description: "Dropbox OAuth 登录；需在 config.toml [cloud.oauth] 配置 dropbox_client_id/secret",
	}, enum.DownloadTaskTypeDropbox.Value(), func(credential string) cloudsync.CloudProvider {
		return NewDropboxProvider(credential)
	})
}

type dropboxEntry struct {
	Tag         string `json:".tag"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	PathDisplay string `json:"path_display"`
	Size        uint64 `json:"size"`
}

type dropboxListResp struct {
	Entries []dropboxEntry `json:"entries"`
	Cursor  string         `json:"cursor"`
	HasMore bool           `json:"has_more"`
}

type DropboxProvider struct {
	session *oauthSession
}

func NewDropboxProvider(credential string) *DropboxProvider {
	return &DropboxProvider{session: newOAuthSession(credential)}
}

func (p *DropboxProvider) SetCredentialUpdateCallback(fn func(string)) {
	p.session.SetCredentialUpdateCallback(fn)
}

func (p *DropboxProvider) ExportCredential() string {
	return p.session.ExportCredential()
}

func (p *DropboxProvider) Name() string { return "dropbox" }

func (p *DropboxProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	var resp struct {
		Name struct {
			DisplayName string `json:"display_name"`
		} `json:"name"`
	}
	if err := p.postJSON(dropboxAPI+"/users/get_current_account", nil, &resp); err != nil {
		return nil, err
	}
	name := resp.Name.DisplayName
	if name == "" {
		name = "Dropbox用户"
	}
	return &cloudsync.CloudUserInfo{Nickname: name}, nil
}

func (p *DropboxProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
    if pdirFid == "" || pdirFid == "0" {
        pdirFid = ""
    }
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}

	all := make([]dropboxEntry, 0)
	body := map[string]interface{}{"path": pdirFid, "recursive": false, "limit": 2000}
	var resp dropboxListResp
	if err := p.postJSON(dropboxAPI+"/files/list_folder", body, &resp); err != nil {
		return nil, 0, err
	}
	all = append(all, resp.Entries...)
	for resp.HasMore {
		contBody := map[string]interface{}{"cursor": resp.Cursor}
		if err := p.postJSON(dropboxAPI+"/files/list_folder/continue", contBody, &resp); err != nil {
			return nil, 0, err
		}
		all = append(all, resp.Entries...)
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
	for _, e := range slice {
		fid := e.PathDisplay
		if fid == "" {
			fid = e.ID
		}
		files = append(files, cloudsync.CloudFile{
			Fid:      fid,
			FileName: e.Name,
			Size:     int64(e.Size),
			IsDir:    e.Tag == "folder",
		})
	}
	return files, len(all), nil
}

func (p *DropboxProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	var resp struct {
		Link string `json:"link"`
		Metadata struct {
			Name string `json:"name"`
			Size uint64 `json:"size"`
		} `json:"metadata"`
	}
	if err := p.postJSON(dropboxAPI+"/files/get_temporary_link", map[string]string{"path": fid}, &resp); err != nil {
		return nil, err
	}
	if resp.Link == "" {
		return nil, fmt.Errorf("未获取到临时下载链接")
	}
	return &cloudsync.CloudDownloadLink{
		DownloadURL: resp.Link,
		FileName:    resp.Metadata.Name,
		Size:        int64(resp.Metadata.Size),
	}, nil
}

func (p *DropboxProvider) postJSON(reqURL string, body interface{}, result interface{}) error {
	cfg, err := oauth.GetProvider("dropbox")
	if err != nil {
		return err
	}
	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := p.session.authorizedRequest(http.MethodPost, reqURL, bytes.NewReader(payload), map[string]string{
			"Content-Type": "application/json",
		})
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			if err := p.session.refresh(cfg.TokenURL, cfg.ClientID, cfg.ClientSecret); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("Dropbox API错误: %d %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		}
		if result != nil {
			return json.Unmarshal(respBody, result)
		}
		return nil
	}
	return fmt.Errorf("Dropbox 请求重试后仍失败")
}
