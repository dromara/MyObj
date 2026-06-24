package uc

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

var _ cloud.CloudProvider = (*UCProvider)(nil)

type UCProvider struct {
	client *Client
}

func NewUCProvider() *UCProvider {
	return &UCProvider{
		client: NewClient(),
	}
}

func (p *UCProvider) Name() string {
	return "uc"
}

func (p *UCProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareToken, err := p.getShareToken(ctx, shareID, pwd)
	if err != nil {
		return nil, err
	}

	files, totalCount, err := p.listFiles(ctx, shareID, shareToken, pwd, "")
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     shareID,
		FileCount: totalCount,
		TotalSize: totalSize,
		Files:     files,
	}, nil
}

func (p *UCProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	files, _, err := p.listFiles(ctx, shareID, shareToken, "", parentFileID)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (p *UCProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	downloadResp, err := p.client.GetDownloadURL(shareID, shareToken, []string{fileID})
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	if len(downloadResp.Data) == 0 {
		return nil, cloud.ErrFileNotFound
	}

	item := downloadResp.Data[0]
	if item.ErrNo != 0 {
		return nil, fmt.Errorf("download error: errno=%d", item.ErrNo)
	}

	return &cloud.DownloadInfo{
		URL:        item.URL,
		Size:       0,
		Expiration: 15 * time.Minute,
		Headers: map[string]string{
			"Referer": "https://drive.uc.cn/",
		},
	}, nil
}

func (p *UCProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

func (p *UCProvider) getShareToken(ctx context.Context, shareID, pwd string) (string, error) {
	tokenResp, err := p.client.GetShareToken(shareID, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "password") || strings.Contains(errMsg, "pwd") {
			return "", cloud.ErrSharePasswordWrong
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") {
			return "", cloud.ErrShareNotFound
		}
		return "", fmt.Errorf("get share token failed: %w", err)
	}

	return tokenResp.St, nil
}

func (p *UCProvider) listFiles(ctx context.Context, shareID, shareToken, pwd, parentFileID string) ([]cloud.ShareFile, int, error) {
	page := 1
	var allFiles []cloud.ShareFile
	totalCount := 0

	for {
		detailResp, err := p.client.GetShareDetail(shareID, shareToken, pwd, page)
		if err != nil {
			return nil, 0, fmt.Errorf("get share detail failed: %w", err)
		}

		if detailResp.Data.Expired {
			return nil, 0, cloud.ErrShareNotFound
		}

		totalCount = detailResp.Data.TotalCount

		for _, item := range detailResp.Data.FileList {
			if parentFileID != "" && item.ParentID != parentFileID {
				continue
			}
			allFiles = append(allFiles, convertFileItem(item))
		}

		if !detailResp.Data.HasMore {
			break
		}
		page++
	}

	return allFiles, totalCount, nil
}

func convertFileItem(item FileItem) cloud.ShareFile {
	updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse(time.RFC3339, item.CreatedAt)
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.FileName,
		Size:         item.FileSize,
		IsDir:        item.IsDir,
		FileType:     getFileType(item.FileType),
		FileExt:      item.FileExtension,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.Thumbnail,
		ParentFileID: item.ParentID,
	}
}

func getFileType(fileType int) string {
	switch fileType {
	case 1:
		return "file"
	case 2:
		return "folder"
	default:
		return "file"
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
	if !strings.Contains(host, "drive.uc.cn") && !strings.Contains(host, "pan.uc.cn") {
		return "", fmt.Errorf("%w: not a valid UC URL", cloud.ErrInvalidShareURL)
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

	return "", cloud.ErrInvalidShareURL
}

func parseUCShareURL(urlStr string) (shareID, pwd string, err error) {
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
