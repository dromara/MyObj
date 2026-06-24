package baidu

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

// 确保 BaiduProvider 实现了 CloudProvider 接口
var _ cloud.CloudProvider = (*BaiduProvider)(nil)

// BaiduProvider 百度网盘提供者
type BaiduProvider struct {
	client *Client
}

// NewBaiduProvider 创建百度网盘提供者
func NewBaiduProvider() *BaiduProvider {
	return &BaiduProvider{
		client: NewClient(),
	}
}

// Name 提供者名称
func (p *BaiduProvider) Name() string {
	return "baidu"
}

// ParseShareLink 解析分享链接
func (p *BaiduProvider) ParseShareLink(ctx context.Context, urlStr, pwd string) (*cloud.ShareInfo, error) {
	// 1. 从URL中提取分享ID
	shareID, err := extractShareID(urlStr)
	if err != nil {
		return nil, err
	}

	// 2. 验证分享链接
	verifyResp, err := p.verifyShare(ctx, shareID, pwd)
	if err != nil {
		return nil, err
	}

	// 3. 列出根目录文件
	files, err := p.listFiles(ctx, verifyResp.ShareID, verifyResp.Uk, "")
	if err != nil {
		return nil, err
	}

	// 4. 计算总大小
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	return &cloud.ShareInfo{
		ShareID:   verifyResp.ShareID,
		Title:     "", // 百度网盘API不直接返回标题
		FileCount: len(files),
		TotalSize: totalSize,
		Files:     files,
		ExpiresAt: nil, // 百度网盘分享链接通常不过期
	}, nil
}

// ListShareFiles 列出分享中的文件
func (p *BaiduProvider) ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]cloud.ShareFile, error) {
	// 需要uk参数，这里简化处理
	// 实际应用中可能需要缓存shareID和uk的映射
	return nil, fmt.Errorf("not implemented: ListShareFiles requires uk parameter")
}

// GetDownloadLink 获取文件下载链接
func (p *BaiduProvider) GetDownloadLink(ctx context.Context, shareID, fileID string) (*cloud.DownloadInfo, error) {
	// 获取uk
	verifyResp, err := p.client.ShareVerify(shareID, "")
	if err != nil {
		return nil, fmt.Errorf("verify share failed: %w", err)
	}

	// 转换fileID为int64
	var fsID int64
	fmt.Sscanf(fileID, "%d", &fsID)

	downloadResp, err := p.client.GetShareDownloadURL(shareID, verifyResp.Uk, fsID)
	if err != nil {
		return nil, fmt.Errorf("get download url failed: %w", err)
	}

	// 优先从顶层获取dlink，否则从info中获取
	dlink := downloadResp.Dlink
	if dlink == "" && len(downloadResp.Info) > 0 {
		dlink = downloadResp.Info[0].Dlink
	}

	if dlink == "" {
		return nil, fmt.Errorf("no download link in response")
	}

	return &cloud.DownloadInfo{
		URL:        dlink,
		Size:       0,
		Expiration: 15 * time.Minute,
		Headers: map[string]string{
			"User-Agent": "pan.baidu.com",
		},
	}, nil
}

// DownloadFile 下载文件
func (p *BaiduProvider) DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error) {
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

// verifyShare 验证分享链接（内部方法）
func (p *BaiduProvider) verifyShare(ctx context.Context, surl, pwd string) (*ShareVerifyResponse, error) {
	verifyResp, err := p.client.ShareVerify(surl, pwd)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "errno=-12") || strings.Contains(errMsg, "提取码错误") {
			return nil, cloud.ErrSharePasswordWrong
		}
		if strings.Contains(errMsg, "errno=-21") || strings.Contains(errMsg, "分享链接不存在") {
			return nil, cloud.ErrShareNotFound
		}
		return nil, fmt.Errorf("verify share failed: %w", err)
	}

	return verifyResp, nil
}

// listFiles 列出文件（内部方法）
func (p *BaiduProvider) listFiles(ctx context.Context, shareID, uk, parentFileID string) ([]cloud.ShareFile, error) {
	fileListResp, err := p.client.ListShareFiles(shareID, uk, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("list share files failed: %w", err)
	}

	files := make([]cloud.ShareFile, 0, len(fileListResp.List))
	for _, item := range fileListResp.List {
		files = append(files, convertFileItem(item))
	}

	return files, nil
}

// convertFileItem 转换文件项
func convertFileItem(item ShareFileItem) cloud.ShareFile {
	// 解析更新时间
	updatedAt := time.Unix(item.ServerMtime, 0)

	return cloud.ShareFile{
		FileID:       fmt.Sprintf("%d", item.FsID),
		Name:         item.Filename,
		Size:         item.Size,
		IsDir:        item.Isdir == 1,
		FileType:     getFileType(item.Category),
		FileExt:      getFileExtension(item.Filename),
		UpdatedAt:    updatedAt,
		Thumbnail:    "",
		ParentFileID: "", // 百度网盘API不返回父目录ID
	}
}

// getFileType 根据类别号获取文件类型
func getFileType(category int) string {
	switch category {
	case 1:
		return "video"
	case 2:
		return "audio"
	case 3:
		return "image"
	case 4:
		return "doc"
	case 5:
		return "app"
	case 6:
		return "torrent"
	case 7:
		return "other"
	default:
		return "other"
	}
}

// getFileExtension 获取文件扩展名
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractShareID 从URL中提取分享ID
func extractShareID(urlStr string) (string, error) {
	// 支持的URL格式:
	// https://pan.baidu.com/s/xxxxx
	// https://pan.baidu.com/share/init?surl=xxxxx
	// https://yun.baidu.com/s/xxxxx
	// xxxx (直接分享ID)

	// 如果是纯ID（不含://），直接返回
	if !strings.Contains(urlStr, "://") {
		// 验证ID格式（字母数字下划线，通常10位以上）
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, nil
		}
		return "", cloud.ErrInvalidShareURL
	}

	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", cloud.ErrInvalidShareURL
	}

	// 检查域名
	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "pan.baidu.com") && !strings.Contains(host, "yun.baidu.com") {
		return "", fmt.Errorf("%w: not a valid baidu pan URL", cloud.ErrInvalidShareURL)
	}

	// 从路径中提取 /s/{shareID}
	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID := parts[i+1]
			// 去除可能的查询参数
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				return shareID, nil
			}
		}
	}

	// 尝试从 surl 参数提取
	surl := parsedURL.Query().Get("surl")
	if surl != "" {
		return surl, nil
	}

	return "", cloud.ErrInvalidShareURL
}

// parseBaiduShareURL 解析百度网盘分享URL（导出函数）
func parseBaiduShareURL(urlStr string) (shareID, pwd string, err error) {
	shareID, err = extractShareID(urlStr)
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