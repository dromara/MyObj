package aliyun

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

// 确保 AliyunProvider 实现了 CloudProvider 接口
var _ cloud.CloudProvider = (*AliyunProvider)(nil)

// AliyunProvider 阿里云盘提供者
type AliyunProvider struct {
	client      *Client
	driveID     string // 从文件列表中缓存的 drive_id
	accessToken string // OAuth2 access_token
}

// SetAccessToken 设置access_token
func (p *AliyunProvider) SetAccessToken(token string) {
	p.accessToken = token
}

// GetAccessToken 获取access_token
func (p *AliyunProvider) GetAccessToken() string {
	return p.accessToken
}

// NewAliyunProvider 创建阿里云盘提供者
func NewAliyunProvider(appID string) *AliyunProvider {
	return &AliyunProvider{
		client: NewClient(appID),
	}
}

// Name 提供者名称
func (p *AliyunProvider) Name() string {
	return "aliyun"
}

// GetDriveID 获取缓存的drive_id
func (p *AliyunProvider) GetDriveID() string {
	return p.driveID
}

// ParseShareLink 解析分享链接
func (p *AliyunProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	// 1. 从URL中提取分享ID和文件夹ID
	shareID, folderID, err := extractShareIDAndFolderID(urlStr)
	if err != nil {
		return nil, err
	}

	// 2. 获取分享Token
	shareToken, err := p.getShareToken(ctx, shareID, pwd)
	if err != nil {
		return nil, err
	}

	// 3. 列出指定目录文件（如果有folderID则列出该目录，否则列出根目录）
	parentID := ""
	if folderID != "" {
		parentID = folderID
	}
	files, err := p.listFiles(ctx, shareID, shareToken, parentID)
	if err != nil {
		return nil, err
	}

	// 4. 计算总大小
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

// ListShareFiles 列出分享中的文件
func (p *AliyunProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	// 获取分享Token（需要缓存优化，这里简化处理）
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}
	
	return p.listFiles(ctx, shareID, shareToken, parentFileID)
}

// GetDownloadLink 获取文件下载链接
func (p *AliyunProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	shareToken, err := p.getShareToken(ctx, shareID, "")
	if err != nil {
		return nil, err
	}

	// 如果没有缓存的drive_id，先获取文件列表
	if p.driveID == "" {
		_, _ = p.listFiles(ctx, shareID, shareToken, "root")
	}

	// 优先使用access_token（OAuth2登录后）
	if p.accessToken != "" && p.driveID != "" {
		downloadResp, err := p.client.GetDownloadURLWithToken(p.accessToken, p.driveID, fileID)
		if err == nil && downloadResp.URL != "" {
			return &cloud.DownloadInfo{
				URL:        downloadResp.URL,
				Size:       downloadResp.Size,
				Expiration: 4 * time.Hour,
				Headers: map[string]string{
					"Referer": "https://www.alipan.com/",
				},
			}, nil
		}
	}

	// 回退到分享token方式
	downloadResp, err := p.client.GetShareLinkDownloadURL(shareID, shareToken, fileID, p.driveID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	// 解析过期时间
	expiration := 15 * time.Minute
	if downloadResp.Expiration != "" {
		d, err := time.ParseDuration(downloadResp.Expiration)
		if err == nil {
			expiration = d
		}
	}

	return &cloud.DownloadInfo{
		URL:        downloadResp.URL,
		Size:       downloadResp.Size,
		Expiration: expiration,
		Headers: map[string]string{
			"Referer": "https://www.alipan.com/",
		},
	}, nil
}

// DownloadFile 下载文件
func (p *AliyunProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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
func (p *AliyunProvider) getShareToken(ctx context.Context, shareID, pwd string) (string, error) {
	tokenResp, err := p.client.GetShareToken(shareID, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "share_link.ExtractCodeError") || 
		   strings.Contains(errMsg, "InvalidParameter.ExtractCode") {
			return "", cloud.ErrSharePasswordWrong
		}
		if strings.Contains(errMsg, "share_link.ShareLinkNotFound") ||
		   strings.Contains(errMsg, "NotFound") {
			return "", cloud.ErrShareNotFound
		}
		return "", fmt.Errorf("get share token failed: %w", err)
	}
	
	return tokenResp.ShareToken, nil
}

// listFiles 列出文件（内部方法）
func (p *AliyunProvider) listFiles(ctx context.Context, shareID, shareToken, parentFileID string) ([]cloud.ShareFile, error) {
	if parentFileID == "" {
		parentFileID = "root"
	}

	fileListResp, err := p.client.ListShareFiles(shareID, shareToken, parentFileID, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(fileListResp.Items))
	for _, item := range fileListResp.Items {
		// 缓存 drive_id（从第一个文件获取）
		if p.driveID == "" && item.DriveID != "" {
			p.driveID = item.DriveID
		}
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

// convertFileItem 转换文件项
func convertFileItem(item ShareFileItem) cloud.ShareFile {
	// 解析更新时间
	updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)
	if updatedAt.IsZero() {
		updatedAt, _ = time.Parse(time.RFC3339, item.CreatedAt)
	}
	
	return cloud.ShareFile{
		FileID:       item.FileID,
		Name:         item.Name,
		Size:         item.Size,
		IsDir:        item.Type == "folder",
		FileType:     item.Category,
		FileExt:      item.FileExtension,
		UpdatedAt:    updatedAt,
		Thumbnail:    item.Thumbnail,
		ParentFileID: item.ParentFileID,
	}
}

// extractShareIDAndFolderID 从URL中提取分享ID和文件夹ID
func extractShareIDAndFolderID(urlStr string) (shareID, folderID string, err error) {
	// 支持的URL格式:
	// https://www.aliyundrive.com/s/xxxxx
	// https://www.alipan.com/s/xxxxx
	// https://www.alipan.com/s/xxxxx/folder/yyyyy
	// xxxx (直接分享ID)

	// 如果是纯ID（不含://），直接返回
	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", cloud.ErrInvalidShareURL
	}

	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", cloud.ErrInvalidShareURL
	}

	// 检查域名
	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "aliyundrive.com") &&
		!strings.Contains(host, "alipan.com") &&
		!strings.Contains(host, "aliyun.com") {
		return "", "", fmt.Errorf("%w: not a valid aliyun drive URL", cloud.ErrInvalidShareURL)
	}

	// 从路径中提取 /s/{shareID}[/folder/{folderID}]
	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID = parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
		}
		if part == "folder" && i+1 < len(parts) {
			folderID = parts[i+1]
			if idx := strings.Index(folderID, "?"); idx > 0 {
				folderID = folderID[:idx]
			}
		}
	}

	if shareID == "" {
		return "", "", cloud.ErrInvalidShareURL
	}
	return shareID, folderID, nil
}

// parseAliyunShareURL 解析阿里云盘分享URL（导出函数）
func parseAliyunShareURL(urlStr string) (shareID, pwd string, err error) {
	shareID, _, err = extractShareIDAndFolderID(urlStr)
	if err != nil {
		return "", "", err
	}

	// 尝试从URL中提取密码
	if strings.Contains(urlStr, "://") {
		parsedURL, parseErr := url.Parse(urlStr)
		if parseErr == nil {
			pwd = parsedURL.Query().Get("pwd")
		}
	}

	return shareID, pwd, nil
}
