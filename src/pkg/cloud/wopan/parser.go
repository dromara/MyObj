package wopan

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"myobj/src/pkg/cloud"
)

var _ cloud.CloudProvider = (*WopanProvider)(nil)

// WopanProvider 联通云盘提供者
type WopanProvider struct {
	client *Client
}

// NewWopanProvider 创建联通云盘提供者
func NewWopanProvider() *WopanProvider {
	return &WopanProvider{
		client: NewClient(),
	}
}

// Name 提供者名称
func (p *WopanProvider) Name() string {
	return "wopan"
}

// ParseShareLink 解析分享链接
func (p *WopanProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	shareToken, err := p.getShareToken(ctx, shareID, pwd)
	if err != nil {
		return nil, err
	}

	shareInfo, err := p.client.GetShareInfo(shareID)
	if err != nil {
		return nil, fmt.Errorf("get share info failed: %w", err)
	}

	if shareInfo.Data.Expired {
		return nil, cloud.ErrShareNotFound
	}

	files, err := p.listFiles(ctx, shareID, shareToken, "")
	if err != nil {
		return nil, err
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var expiresAt *time.Time
	if shareInfo.Data.Expiration != "" {
		t, err := time.Parse(time.RFC3339, shareInfo.Data.Expiration)
		if err == nil {
			expiresAt = &t
		}
	}

	title := shareInfo.Data.ShareTitle
	if title == "" {
		title = shareInfo.Data.ShareName
	}

	return &cloud.ShareInfo{
		ShareID:   shareID,
		Title:     title,
		FileCount: shareInfo.Data.FileCount + shareInfo.Data.FolderCount,
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: expiresAt,
	}, nil
}

// ListShareFiles 列出分享中的文件
func (p *WopanProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	return p.listFiles(ctx, shareID, shareToken, parentFileID)
}

// GetDownloadLink 获取文件下载链接
func (p *WopanProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	downloadResp, err := p.client.GetDownloadURL(shareID, shareToken, fileID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	expiration := 15 * time.Minute
	if downloadResp.Data.Expiration > 0 {
		expiration = time.Duration(downloadResp.Data.Expiration) * time.Second
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.Data.URL,
		Size:       downloadResp.Data.Size,
		Expiration: expiration,
		Headers: map[string]string{
			"Referer": "https://pan.wo.cn/",
		},
	}, nil
}

// DownloadFile 下载文件
func (p *WopanProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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
func (p *WopanProvider) getShareToken(ctx context.Context, shareID, pwd string) (string, error) {
	if pwd == "" {
		return "", nil
	}

	verifyResp, err := p.client.VerifySharePwd(shareID, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "password") || strings.Contains(errMsg, "pwd") {
			return "", cloud.ErrSharePasswordWrong
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "expired") {
			return "", cloud.ErrShareNotFound
		}
		return "", fmt.Errorf("verify share password failed: %w", err)
	}

	if !verifyResp.Data.Valid {
		return "", cloud.ErrSharePasswordWrong
	}

	return verifyResp.Data.ShareToken, nil
}

// listFiles 列出文件（内部方法）
func (p *WopanProvider) listFiles(ctx context.Context, shareID, shareToken, parentFileID string) ([]cloud.ShareFile, error) {
	fileListResp, err := p.client.GetShareFileList(shareID, shareToken, parentFileID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(fileListResp.Data.Items))
	for _, item := range fileListResp.Data.Items {
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

// convertFileItem 转换文件项
func convertFileItem(item ShareFileItem) cloud.ShareFile {
	updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse(time.RFC3339, item.CreatedAt)
	}

	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.Name,
		Size:         item.Size,
		IsDir:        item.Type == 2,
		FileType:     item.FileType,
		FileExt:      item.FileExt,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.Thumbnail,
		ParentFileID: item.ParentFileID,
	}
}

// extractShareID 从URL中提取分享ID
func extractShareID(urlStr string) (string, error) {
	shareID, _, err := extractShareIDAndPwd(urlStr)
	return shareID, err
}

// extractShareIDAndPwd 从URL中提取分享ID和密码
func extractShareIDAndPwd(urlStr string) (shareID, pwd string, err error) {
	shareID, pwd, err = cloud.ParseShareURL(cloud.ProviderWopan, urlStr)
	return shareID, pwd, err
}
