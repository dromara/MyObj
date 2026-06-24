package xunlei

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

var _ cloud.CloudProvider = (*XunleiProvider)(nil)

type XunleiProvider struct {
	client *Client
}

func NewXunleiProvider() *XunleiProvider {
	return &XunleiProvider{
		client: NewClient(),
	}
}

func (p *XunleiProvider) Name() string {
	return "xunlei"
}

func (p *XunleiProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	linkInfo, err := p.client.GetShareLinkInfo(shareID, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "password") || strings.Contains(errMsg, "pwd") {
			return nil, cloud.ErrSharePasswordRequired
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") {
			return nil, cloud.ErrShareNotFound
		}
		return nil, fmt.Errorf("get share link info failed: %w", err)
	}

	files, err := p.listFiles(ctx, shareID, "", pwd)
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var expiresAt *time.Time
	if linkInfo.ExpireTime > 0 {
		t := time.Unix(linkInfo.ExpireTime, 0)
		expiresAt = &t
	}

	title := linkInfo.Title
	if title == "" {
		title = linkInfo.ShareTitle
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     title,
		FileCount: linkInfo.FileCount,
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

func (p *XunleiProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	return p.listFiles(ctx, shareID, parentFileID, "")
}

func (p *XunleiProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	downloadResp, err := p.client.GetShareDownload(shareID, fileID, "")
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	expiration := 15 * time.Minute
	if downloadResp.ExpireTime > 0 {
		d := time.Until(time.Unix(downloadResp.ExpireTime, 0))
		if d > 0 {
			expiration = d
		}
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.URL,
		Size:       downloadResp.Size,
		Expiration: expiration,
		Headers:    map[string]string{},
	}, nil
}

func (p *XunleiProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

func (p *XunleiProvider) listFiles(ctx context.Context, shareID, parentFileID, pwd string) ([]cloud.ShareFile, error) {
	fileListResp, err := p.client.ListShareFiles(shareID, parentFileID, pwd, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(fileListResp.Files))
	for _, item := range fileListResp.Files {
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

func convertFileItem(item ShareFileItem) cloud.ShareFile {
	updatedAt := time.Unix(item.UpdateTime, 0)
	if updatedAt.IsZero() {
		updatedAt = time.Unix(item.CreateTime, 0)
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.Name,
		Size:         item.Size,
		IsDir:        item.IsDir,
		FileType:     item.FileType,
		FileExt:      item.FileExt,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.Thumbnail,
		ParentFileID: item.ParentFileID,
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
	if !strings.Contains(host, "xunlei.com") {
		return "", fmt.Errorf("%w: not a valid xunlei URL", cloud.ErrInvalidShareURL)
	}

	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID := parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				return shareID, nil
			}
		}
	}

	shareID := parsedURL.Query().Get("share_id")
	if shareID != "" {
		return shareID, nil
	}

	return "", cloud.ErrInvalidShareURL
}

func parseXunleiShareURL(urlStr string) (shareID, pwd string, err error) {
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
