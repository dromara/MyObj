package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/cloudsync/oauth"
	"myobj/src/pkg/enum"
)

const googleDriveAPI = "https://www.googleapis.com/drive/v3"

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID:          "google",
		Name:        "Google Drive",
		AuthType:    cloudsync.AuthOAuth2,
		Description: "Google Drive OAuth 登录；需在 config.toml [cloud.oauth] 配置 google_client_id/secret",
	}, enum.DownloadTaskTypeGoogleDrive.Value(), func(credential string) cloudsync.CloudProvider {
		return NewGoogleDriveProvider(credential)
	})
}

type googleDriveFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     string `json:"size"`
}

type googleDriveListResp struct {
	Files         []googleDriveFile `json:"files"`
	NextPageToken string            `json:"nextPageToken"`
}

type GoogleDriveProvider struct {
	session *oauthSession
}

func NewGoogleDriveProvider(credential string) *GoogleDriveProvider {
	return &GoogleDriveProvider{session: newOAuthSession(credential)}
}

func (p *GoogleDriveProvider) SetCredentialUpdateCallback(fn func(string)) {
	p.session.SetCredentialUpdateCallback(fn)
}

func (p *GoogleDriveProvider) ExportCredential() string {
	return p.session.ExportCredential()
}

func (p *GoogleDriveProvider) Name() string { return "google" }

func (p *GoogleDriveProvider) Validate() (*cloudsync.CloudUserInfo, error) {
	var resp struct {
		User struct {
			DisplayName string `json:"displayName"`
		} `json:"user"`
	}
	if err := p.getJSON(googleDriveAPI+"/about?fields=user", &resp); err != nil {
		return nil, err
	}
	name := resp.User.DisplayName
	if name == "" {
		name = "Google Drive用户"
	}
	return &cloudsync.CloudUserInfo{Nickname: name}, nil
}

func (p *GoogleDriveProvider) ListFiles(pdirFid string, page, size int) ([]cloudsync.CloudFile, int, error) {
	if pdirFid == "" {
		pdirFid = "root"
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 50
	}

	q := url.QueryEscape(fmt.Sprintf("'%s' in parents and trashed = false", pdirFid))
	all := make([]googleDriveFile, 0)
	pageToken := ""
	for {
		reqURL := googleDriveAPI + "/files?q=" + q + "&pageSize=1000&fields=files(id,name,mimeType,size),nextPageToken&orderBy=folder,name&includeItemsFromAllDrives=true&supportsAllDrives=true"
		if pageToken != "" {
			reqURL += "&pageToken=" + url.QueryEscape(pageToken)
		}
		var resp googleDriveListResp
		if err := p.getJSON(reqURL, &resp); err != nil {
			return nil, 0, err
		}
		all = append(all, resp.Files...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
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
		var sz int64
		fmt.Sscanf(f.Size, "%d", &sz)
		files = append(files, cloudsync.CloudFile{
			Fid:      f.ID,
			FileName: f.Name,
			Size:     sz,
			IsDir:    f.MimeType == "application/vnd.google-apps.folder",
		})
	}
	return files, len(all), nil
}

func (p *GoogleDriveProvider) GetDownloadLink(fid string) (*cloudsync.CloudDownloadLink, error) {
	var meta googleDriveFile
	if err := p.getJSON(googleDriveAPI+"/files/"+url.PathEscape(fid)+"?fields=id,name,size,mimeType", &meta); err != nil {
		return nil, err
	}
	if meta.MimeType == "application/vnd.google-apps.folder" {
		return nil, fmt.Errorf("不能下载文件夹")
	}
	downloadURL := googleDriveAPI + "/files/" + url.PathEscape(fid) + "?alt=media&supportsAllDrives=true"
	var sz int64
	fmt.Sscanf(meta.Size, "%d", &sz)
	return &cloudsync.CloudDownloadLink{
		DownloadURL: downloadURL,
		FileName:    meta.Name,
		Size:        sz,
		Headers: map[string]string{
			"Authorization": "Bearer " + p.session.access(),
		},
	}, nil
}

func (p *GoogleDriveProvider) getJSON(reqURL string, result interface{}) error {
	cfg, err := oauth.GetProvider("google")
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
			return fmt.Errorf("Google Drive API错误: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		if result != nil {
			return json.Unmarshal(body, result)
		}
		return nil
	}
	return fmt.Errorf("Google Drive 请求重试后仍失败")
}
