package caiyun

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"myobj/src/pkg/cloud"
)

var _ cloud.CloudProvider = (*CaiyunProvider)(nil)

type CaiyunProvider struct {
	client *Client
}

func NewCaiyunProvider() *CaiyunProvider {
	return &CaiyunProvider{
		client: NewClient(),
	}
}

func (p *CaiyunProvider) Name() string {
	return "caiyun"
}

func (p *CaiyunProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareInfoResp, err := p.client.GetShareInfo(shareID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "不存在") {
			return nil, cloud.ErrShareNotFound
		}
		return nil, fmt.Errorf("get share info failed: %w", err)
	}

	accessCode := shareInfoResp.Data.AccessCode
	if pwd != "" {
		accessCode = pwd
	}

	files, err := p.listFiles(ctx, shareID, accessCode, "")
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "密码") || strings.Contains(errMsg, "password") ||
			strings.Contains(errMsg, "accessCode") || strings.Contains(errMsg, "提取码") {
			return nil, cloud.ErrSharePasswordWrong
		}
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var expiresAt *time.Time
	if shareInfoResp.Data.ExpireTime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", shareInfoResp.Data.ExpireTime)
		if err != nil {
			t, err = time.Parse(time.RFC3339, shareInfoResp.Data.ExpireTime)
		}
		if err == nil {
			expiresAt = &t
		}
	}

	title := shareInfoResp.Data.ShareTitle
	if title == "" {
		title = shareInfoResp.Data.ShareName
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     title,
		FileCount: shareInfoResp.Data.FileCount + shareInfoResp.Data.FolderCount,
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

func (p *CaiyunProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	return p.listFiles(ctx, shareID, "", parentFileID)
}

func (p *CaiyunProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	downloadResp, err := p.client.GetShareDownloadURL(shareID, "", fileID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	if downloadResp.Data.DownloadURL == "" {
		return nil, fmt.Errorf("empty download url returned")
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.Data.DownloadURL,
		Size:       downloadResp.Data.FileSize,
		Expiration: 15 * time.Minute,
		Headers: map[string]string{
			"Referer": "https://caiyun.139.com/",
		},
	}, nil
}

func (p *CaiyunProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

func (p *CaiyunProvider) listFiles(ctx context.Context, shareID, accessCode, catalogID string) ([]cloud.ShareFile, error) {
	var allFiles []cloud.ShareFile
	pageNum := 1
	pageSize := 200

	for {
		resp, err := p.client.GetShareFileList(shareID, accessCode, catalogID, pageNum, pageSize)
		if err != nil {
			return nil, fmt.Errorf("list share files failed: %w", err)
		}

		for _, item := range resp.Data.FileList {
			allFiles = append(allFiles, convertFileItem(item))
		}

		for _, catalog := range resp.Data.CatalogList {
			allFiles = append(allFiles, cloud.ShareFile{
				FileID:   catalog.CatalogID,
				Name:     catalog.CatalogName,
				Size:     0,
				IsDir:    true,
				FileType: "folder",
			})
		}

		if len(resp.Data.FileList)+len(resp.Data.CatalogList) < pageSize {
			break
		}
		if resp.Data.TotalCount > 0 && len(allFiles) >= resp.Data.TotalCount {
			break
		}
		pageNum++
	}

	return allFiles, nil
}

func convertFileItem(item ShareFileItem) cloud.ShareFile {
	updatedAt, _ := time.Parse("2006-01-02 15:04:05", item.UpdateTime)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse("2006-01-02 15:04:05", item.CreateTime)
	}

	fileExt := ""
	if idx := strings.LastIndex(item.FileName, "."); idx >= 0 {
		fileExt = item.FileName[idx+1:]
	}

	return cloud.ShareFile{
		FileID:    item.FileID,
		Name:      item.FileName,
		Size:      item.FileSize,
		IsDir:     false,
		FileType:  item.FileType,
		FileExt:   fileExt,
		UpdatedAt: updatedAt,
		Thumbnail: item.ThumbnailURL,
	}
}

func extractShareID(urlStr string) (string, error) {
	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, nil
		}
		return "", cloud.ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", cloud.ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "caiyun.139.com") && !strings.Contains(host, "caiyun.com") {
		return "", fmt.Errorf("%w: not a valid caiyun URL", cloud.ErrInvalidShareURL)
	}

	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if (part == "s" || part == "w") && i+1 < len(parts) {
			shareID := parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				return shareID, nil
			}
		}
	}

	return "", cloud.ErrInvalidShareURL
}

func parseCaiyunShareURL(urlStr string) (shareID, pwd string, err error) {
	shareID, err = extractShareID(urlStr)
	if err != nil {
		return "", "", err
	}

	if strings.Contains(urlStr, "://") {
		parsedURL, parseErr := url.Parse(urlStr)
		if parseErr == nil {
			pwd = parsedURL.Query().Get("pwd")
		}
	}

	return shareID, pwd, nil
}
