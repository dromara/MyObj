package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/oauth"
	"myobj/src/pkg/enum"
)

const oneDriveAPI = "https://graph.microsoft.com/v1.0/me/drive"

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "onedrive",
		Name:        "OneDrive",
		AuthType:    cloudsync.AuthOAuth2,
		Description: "Microsoft OneDrive OAuth 登录；需在 config.toml [cloud.oauth] 配置 onedrive_client_id/secret",
	}, enum.DownloadTaskTypeOneDrive.Value(), func(credential string) cloudsync.CloudProvider {
		return NewOneDriveProvider(credential)
	})
}

type oneDriveItem struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Size                 int64  `json:"size"`
	DownloadURL          string `json:"@microsoft.graph.downloadUrl"`
	File                 *struct{} `json:"file"`
	Folder               *struct{} `json:"folder"`
}

type oneDriveListResp struct {
	Value    []oneDriveItem `json:"value"`
	NextLink string         `json:"@odata.nextLink"`
}

type OneDriveProvider struct {
	session *oauthSession
}

func NewOneDriveProvider(credential string) *OneDriveProvider {
	return &OneDriveProvider{session: newOAuthSession(credential)}
}

func (p *OneDriveProvider) SetCredentialUpdateCallback(fn func(string)) {
	p.session.SetCredentialUpdateCallback(fn)
}

func (p *OneDriveProvider) ExportCredential() string {
	return p.session.ExportCredential()
}

func (p *OneDriveProvider) Name() string { return "onedrive" }

func (p *OneDriveProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	var resp struct {
		DisplayName string `json:"displayName"`
	}
	if err := p.getJSON("https://graph.microsoft.com/v1.0/me", &resp); err != nil {
		return nil, err
	}
	name := resp.DisplayName
	if name == "" {
		name = "OneDrive用户"
	}
	return &cloudsync.CloudUserInfo{Nickname: name}, nil
}

func (p *OneDriveProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if pdirFid == "" {
		pdirFid = "root"
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}

	reqURL := oneDriveAPI + "/root/children?$top=1000&$select=id,name,size,file,folder"
	if pdirFid != "root" {
		reqURL = oneDriveAPI + "/items/" + url.PathEscape(pdirFid) + "/children?$top=1000&$select=id,name,size,file,folder"
	}

	all := make([]oneDriveItem, 0)
	next := reqURL
	for next != "" {
		var resp oneDriveListResp
		if err := p.getJSONURL(next, &resp); err != nil {
			return nil, 0, err
		}
		all = append(all, resp.Value...)
		next = resp.NextLink
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
	for _, item := range slice {
		files = append(files, cloudsync.CloudFile{
			Fid:      item.ID,
			FileName: item.Name,
			Size:     item.Size,
			IsDir:    item.Folder != nil,
		})
	}
	return files, len(all), nil
}

func (p *OneDriveProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	var item oneDriveItem
	path := oneDriveAPI + "/items/" + url.PathEscape(fid) + "?$select=id,name,size,@microsoft.graph.downloadUrl,file"
	if err := p.getJSON(path, &item); err != nil {
		return nil, err
	}
	if item.DownloadURL == "" {
		return nil, fmt.Errorf("未获取到下载链接")
	}
	return &cloudsync.CloudDownloadLink{
		DownloadURL: item.DownloadURL,
		FileName:    item.Name,
		Size:        item.Size,
	}, nil
}

func (p *OneDriveProvider) getJSON(path string, result interface{}) error {
	return p.getJSONURL(path, result)
}

func (p *OneDriveProvider) getJSONURL(reqURL string, result interface{}) error {
	cfg, err := oauth.GetProvider("onedrive")
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := p.session.authorizedRequest(http.MethodGet, reqURL, nil, map[string]string{
			"Accept": "application/json",
		})
		if err != nil {
			return fmt.Errorf("请求失败: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized && attempt == 0 {
			if err := p.session.refresh(cfg.TokenURL, cfg.ClientID, cfg.ClientSecret); err != nil {
				return err
			}
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("OneDrive API错误: %d %s", resp.StatusCode, string(body))
		}
		if result != nil {
			return json.Unmarshal(body, result)
		}
		return nil
	}
	return fmt.Errorf("OneDrive 请求重试后仍失败")
}
