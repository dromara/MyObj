package tianyi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"myobj/src/pkg/cloud"
)

var ErrTianyiAPIError = fmt.Errorf("tianyi api error")

var _ cloud.CloudProvider = (*TianyiProvider)(nil)

type TianyiProvider struct {
	client *Client
}

func NewTianyiProvider() *TianyiProvider {
	return &TianyiProvider{
		client: NewClient(),
	}
}

func (p *TianyiProvider) Name() string {
	return "tianyi"
}

func (p *TianyiProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareInfo, err := p.client.GetShareInfo(shareID)
	if err != nil {
		return nil, fmt.Errorf("get share info failed: %w", err)
	}

	if shareInfo.Expired {
		return nil, cloud.ErrShareNotFound
	}

	if shareInfo.NeedAccessCode {
		if pwd == "" {
			return nil, cloud.ErrSharePasswordRequired
		}
		loginResp, err := p.client.CheckAccessCode(shareID, pwd)
		if err != nil {
			return nil, fmt.Errorf("check access code failed: %w", err)
		}
		if loginResp.ValidationStatus != 1 {
			return nil, cloud.ErrSharePasswordWrong
		}
	}

	files, err := p.listFiles(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var expiresAt *time.Time
	if shareInfo.ExpireTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", shareInfo.ExpireTime)
		if err == nil {
			expiresAt = &t
		}
	}

	title := shareInfo.ShareTitle
	if title == "" {
		title = shareInfo.FileName
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     title,
		FileCount: len(files),
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

func (p *TianyiProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	return p.listFiles(ctx, shareID, parentFileID)
}

func (p *TianyiProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	downloadResp, err := p.client.GetShareDownloadURL(shareID, fileID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.DownloadURL,
		Size:       downloadResp.FileSize,
		Expiration: 15 * time.Minute,
		Headers: map[string]string{
			"Referer": "https://cloud.189.cn/",
		},
	}, nil
}

func (p *TianyiProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
	downloadInfo, err := p.GetDownloadLink(ctx, shareID, fileID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadInfo.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do download request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: status=%d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (p *TianyiProvider) listFiles(ctx context.Context, shareID, parentFolderID string) ([]cloud.ShareFile, error) {
	fileListResp, err := p.client.ListShareDir(shareID, parentFolderID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("list share dir failed: %w", err)
	}

	fileList := fileListResp.FileListAO.FileList
	files := make([]cloud.ShareFile, 0, len(fileList))
	for _, item := range fileList {
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

func convertFileItem(item ShareFileDTO) cloud.ShareFile {
	updatedAt, _ := time.Parse("2006-01-02 15:04:05", item.LastOpTime)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse("2006-01-02 15:04:05", item.CreateDate)
	}

	var fileExt string
	if !item.IsFolder {
		fileExt = strings.TrimPrefix(filepath.Ext(item.FileName), ".")
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.FileName,
		Size:         item.FileSize,
		IsDir:        item.IsFolder,
		FileType:     item.FileType,
		FileExt:      fileExt,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.SmallIconURL,
		ParentFileID: item.ParentFolderID,
	}
}

func extractShareID(urlStr string) (string, error) {
	shareID, _, err := cloud.ParseShareURL(cloud.ProviderTianyi, urlStr)
	if err != nil {
		return "", err
	}
	return shareID, nil
}
