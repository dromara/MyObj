package quark

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

// 确保 QuarkProvider 实现了 CloudProvider 接口
var _ cloud.CloudProvider = (*QuarkProvider)(nil)

// QuarkProvider 夸克网盘提供者
type QuarkProvider struct {
	client *Client
}

// NewQuarkProvider 创建夸克网盘提供者
func NewQuarkProvider() *QuarkProvider {
	return &QuarkProvider{
		client: NewClient(),
	}
}

// Name 提供者名称
func (p *QuarkProvider) Name() string {
	return "quark"
}

// ParseShareLink 解析分享链接
func (p *QuarkProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareToken, err := p.getShareToken(ctx, shareID, pwd)
	if err != nil {
		return nil, err
	}

	detail, err := p.client.GetShareDetail(shareID, shareToken, pwd, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("get share detail failed: %w", err)
	}

	if detail.Expired {
		return nil, cloud.ErrShareNotFound
	}

	files := make([]cloud.ShareFile, 0, len(detail.List))
	for _, item := range detail.List {
		files = append(files, convertFileItem(item))
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var expiresAt *time.Time
	if detail.ExpireTime > 0 {
		t := time.Unix(detail.ExpireTime, 0)
		expiresAt = &t
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     detail.Title,
		FileCount: detail.Count,
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

// ListShareFiles 列出分享中的文件
func (p *QuarkProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	detail, err := p.client.GetShareDetail(shareID, shareToken, "", 1, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(detail.List))
	for _, item := range detail.List {
		if parentFileID == "" || item.ParentFileID == parentFileID {
			files = append(files, convertFileItem(item))
		}
	}

	return files, nil
}

// GetDownloadLink 获取文件下载链接
func (p *QuarkProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	tokenResp, err := p.client.GetDownloadToken(shareID, shareToken, []string{fileID})
	if err != nil {
		return nil, fmt.Errorf("get download token failed: %w", err)
	}

	if tokenResp.DownloadToken == "" {
		return nil, fmt.Errorf("empty download token")
	}

	downloadResp, err := p.client.GetDownloadURL(shareID, shareToken, tokenResp.DownloadToken, fileID, "")
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.DownloadURL,
		Size:       downloadResp.Size,
		Expiration: 15 * time.Minute,
		Headers: map[string]string{
			"Referer": "https://pan.quark.cn/",
		},
	}, nil
}

// DownloadFile 下载文件
func (p *QuarkProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

// getShareToken 获取分享Token（内部方法）
func (p *QuarkProvider) getShareToken(ctx context.Context, shareID, pwd string) (string, error) {
	tokenResp, err := p.client.GetShareToken(shareID, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "password") || strings.Contains(errMsg, "passcode") {
			return "", cloud.ErrSharePasswordWrong
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") {
			return "", cloud.ErrShareNotFound
		}
		return "", fmt.Errorf("get share token failed: %w", err)
	}

	if tokenResp.ShareToken != "" {
		return tokenResp.ShareToken, nil
	}
	if tokenResp.Stoken != "" {
		return tokenResp.Stoken, nil
	}

	return "", fmt.Errorf("empty share token in response")
}

// convertFileItem 转换文件项
func convertFileItem(item ShareFileItem) cloud.ShareFile {
	updatedAt := time.Unix(item.UpdatedAt, 0)
	if updatedAt.IsZero() {
		updatedAt = time.Unix(item.CreatedAt, 0)
	}

	fileType := "file"
	if item.Dir {
		fileType = "folder"
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.Name,
		Size:         item.Size,
		IsDir:        item.Dir,
		FileType:     fileType,
		FileExt:      item.FormatType,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.Thumbnail,
		ParentFileID: item.ParentFileID,
	}
}

// extractShareID 从URL中提取分享ID
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
	if !strings.Contains(host, "quark.cn") {
		return "", fmt.Errorf("%w: not a valid quark URL", cloud.ErrInvalidShareURL)
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

// parseQuarkShareURL 解析夸克网盘分享URL（导出函数）
func parseQuarkShareURL(urlStr string) (shareID, pwd string, err error) {
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
