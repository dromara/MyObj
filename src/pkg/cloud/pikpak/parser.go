package pikpak

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

var _ cloud.CloudProvider = (*PikPakProvider)(nil)

type PikPakProvider struct {
	client *Client
}

func NewPikPakProvider() *PikPakProvider {
	return &PikPakProvider{
		client: NewClient(),
	}
}

func (p *PikPakProvider) Name() string {
	return "pikpak"
}

func (p *PikPakProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareToken, err := p.getShareToken(ctx, shareID)
	if err != nil {
		return nil, err
	}

	files, err := p.listFiles(ctx, shareID, shareToken, "")
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     "",
		FileCount: len(files),
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: nil,
	}, nil
}

func (p *PikPakProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	shareToken, err := p.getShareToken(ctx, shareID)
	if err != nil {
		return nil, err
	}

	return p.listFiles(ctx, shareID, shareToken, parentFileID)
}

func (p *PikPakProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	shareToken, err := p.getShareToken(ctx, shareID)
	if err != nil {
		return nil, err
	}

	downloadResp, err := p.client.GetDownloadURL(shareToken, fileID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	downloadURL := downloadResp.WebContentLink
	if downloadURL == "" {
		downloadURL = downloadResp.Link
	}

	return &cloud.DownloadInfo{
		URL:        downloadURL,
		Size:       downloadResp.Size,
		Expiration: 15 * time.Minute,
		Headers:    map[string]string{},
	}, nil
}

func (p *PikPakProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

func (p *PikPakProvider) getShareToken(ctx context.Context, shareID string) (string, error) {
	tokenResp, err := p.client.GetShareToken(shareID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") {
			return "", cloud.ErrShareNotFound
		}
		if strings.Contains(errMsg, "password") {
			return "", cloud.ErrSharePasswordRequired
		}
		return "", fmt.Errorf("get share token failed: %w", err)
	}

	return tokenResp.ShareToken, nil
}

func (p *PikPakProvider) listFiles(ctx context.Context, shareID, shareToken, parentID string) ([]cloud.ShareFile, error) {
	detailResp, err := p.client.GetShareDetail(shareID, shareToken, parentID, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(detailResp.Files))
	for _, item := range detailResp.Files {
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

func convertFileItem(item ShareFileItem) cloud.ShareFile {
	updatedAt, _ := time.Parse(time.RFC3339, item.ModifiedTime)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse(time.RFC3339, item.CreatedTime)
	}

	return cloud.ShareFile{
		FileID:    item.FileID,
		Name:      item.Name,
		Size:      item.Size,
		IsDir:     item.FolderType == "folder" || item.Kind == "drive#folder",
		FileType:  item.MimeType,
		FileExt:   getFileExtension(item.Name),
		UpdatedAt: updatedAt,
		Thumbnail: item.ThumbnailLink,
	}
}

func getFileExtension(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
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
	if !strings.Contains(host, "mypikpak.com") && !strings.Contains(host, "pikpak.com") {
		return "", fmt.Errorf("%w: not a valid pikpak URL", cloud.ErrInvalidShareURL)
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
