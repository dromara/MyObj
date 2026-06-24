package cloud

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ShareInfo 分享链接信息
type ShareInfo struct {
	ShareID   string       `json:"share_id"`   // 分享ID
	Title     string       `json:"title"`      // 分享标题
	FileCount int          `json:"file_count"` // 文件数量
	TotalSize int64        `json:"total_size"` // 总大小
	Files     []ShareFile  `json:"files"`      // 文件列表
	ExpiresAt *time.Time   `json:"expires_at"` // 过期时间
}

// ShareFile 分享文件信息
type ShareFile struct {
	FileID       string    `json:"file_id"`       // 文件ID
	Name         string    `json:"name"`          // 文件名
	Size         int64     `json:"size"`          // 文件大小
	IsDir        bool      `json:"is_dir"`        // 是否是目录
	FileType     string    `json:"file_type"`     // 文件类型
	FileExt      string    `json:"file_ext"`      // 文件扩展名
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
	Thumbnail    string    `json:"thumbnail"`     // 缩略图URL
	ParentFileID string    `json:"parent_file_id"` // 父目录ID
}

// DownloadInfo 文件下载信息
type DownloadInfo struct {
	URL        string            `json:"url"`         // 下载链接
	Size       int64             `json:"size"`        // 文件大小
	Expiration time.Duration     `json:"expiration"`  // 链接过期时间
	Headers    map[string]string `json:"headers"`     // 下载时需要的请求头
}

// CloudProvider 云盘提供者接口
type CloudProvider interface {
	// Name 提供者名称
	Name() string

	// ParseShareLink 解析分享链接
	// url: 分享链接
	// pwd: 提取码（可选）
	// 返回分享信息
	ParseShareLink(ctx context.Context, url, pwd string) (*ShareInfo, error)

	// ListShareFiles 列出分享中的文件
	// shareID: 分享ID
	// parentFileID: 父目录ID（为空则列出根目录）
	ListShareFiles(ctx context.Context, shareID, parentFileID string) ([]ShareFile, error)

	// GetDownloadLink 获取文件下载链接
	// shareID: 分享ID
	// fileID: 文件ID
	GetDownloadLink(ctx context.Context, shareID, fileID string) (*DownloadInfo, error)

	// DownloadFile 下载文件
	// shareID: 分享ID
	// fileID: 文件ID
	// 返回文件流
	DownloadFile(ctx context.Context, shareID, fileID string) (io.ReadCloser, error)
}

// ProviderType 云盘提供者类型
type ProviderType string

const (
	ProviderAliyun   ProviderType = "aliyun"   // 阿里云盘
	ProviderBaidu    ProviderType = "baidu"    // 百度网盘
	ProviderXunlei   ProviderType = "xunlei"   // 迅雷网盘
	Provider115      ProviderType = "115"      // 115网盘
	ProviderQuark    ProviderType = "quark"    // 夸克网盘
	ProviderPikPak   ProviderType = "pikpak"   // PikPak
	ProviderThunder  ProviderType = "thunder"  // 迅雷
	ProviderCaiyun   ProviderType = "caiyun"   // 移动云盘(和彩云)
	ProviderTianyi   ProviderType = "tianyi"   // 天翼云盘
	ProviderUC       ProviderType = "uc"       // UC网盘
	ProviderWopan    ProviderType = "wopan"    // 联通云盘
)

// ProviderInfo 云盘提供者信息
type ProviderInfo struct {
	Type        ProviderType `json:"type"`
	Name        string       `json:"name"`
	Icon        string       `json:"icon"`
	Enabled     bool         `json:"enabled"`
	Description string       `json:"description"`
}

// GetSupportedProviders 获取支持的云盘提供者列表
func GetSupportedProviders() []ProviderInfo {
	return []ProviderInfo{
		{ProviderAliyun, "阿里云盘", "aliyun", true, "支持阿里云盘分享链接解析"},
		{ProviderBaidu, "百度网盘", "baidu", true, "支持百度网盘分享链接解析"},
		{ProviderXunlei, "迅雷网盘", "xunlei", true, "支持迅雷网盘分享链接解析"},
		{Provider115, "115网盘", "115", true, "支持115网盘分享链接解析"},
		{ProviderQuark, "夸克网盘", "quark", true, "支持夸克网盘分享链接解析"},
		{ProviderPikPak, "PikPak", "pikpak", true, "支持PikPak分享链接解析"},
		{ProviderThunder, "迅雷", "thunder", true, "支持迅雷分享链接解析"},
		{ProviderCaiyun, "和彩云", "caiyun", true, "支持中国移动和彩云分享链接解析"},
		{ProviderTianyi, "天翼云盘", "tianyi", true, "支持中国电信天翼云盘分享链接解析"},
		{ProviderUC, "UC网盘", "uc", true, "支持UC网盘分享链接解析"},
		{ProviderWopan, "联通云盘", "wopan", true, "支持中国联通云盘分享链接解析"},
	}
}

// ParseShareURL 从分享链接中提取分享ID和密码
func ParseShareURL(provider ProviderType, url string) (shareID, pwd string, err error) {
	switch provider {
	case ProviderAliyun:
		return parseAliyunShareURL(url)
	case ProviderBaidu:
		return parseBaiduShareURL(url)
	case ProviderXunlei, ProviderThunder:
		return parseXunleiShareURL(url)
	case Provider115:
		return parse115ShareURL(url)
	case ProviderQuark:
		return parseQuarkShareURL(url)
	case ProviderCaiyun:
		return parseCaiyunShareURL(url)
	case ProviderTianyi:
		return parseTianyiShareURL(url)
	case ProviderUC:
		return parseUCShareURL(url)
	case ProviderWopan:
		return parseWopanShareURL(url)
	case ProviderPikPak:
		return parsePikPakShareURL(url)
	default:
		return "", "", ErrUnsupportedProvider
	}
}

// DetectProvider 从URL自动检测云盘提供者
func DetectProvider(urlStr string) (ProviderType, error) {
	lowerURL := strings.ToLower(urlStr)
	switch {
	case strings.Contains(lowerURL, "aliyundrive.com") || strings.Contains(lowerURL, "alipan.com"):
		return ProviderAliyun, nil
	case strings.Contains(lowerURL, "pan.baidu.com") || strings.Contains(lowerURL, "yun.baidu.com"):
		return ProviderBaidu, nil
	case strings.Contains(lowerURL, "pan.xunlei.com") || strings.Contains(lowerURL, "xunlei.com"):
		return ProviderXunlei, nil
	case strings.Contains(lowerURL, "115.com") || strings.Contains(lowerURL, "115cdn.com"):
		return Provider115, nil
	case strings.Contains(lowerURL, "pan.quark.cn") || strings.Contains(lowerURL, "quark.cn"):
		return ProviderQuark, nil
	case strings.Contains(lowerURL, "mypikpak.com") || strings.Contains(lowerURL, "pikpak.com"):
		return ProviderPikPak, nil
	case strings.Contains(lowerURL, "caiyun.139.com") || strings.Contains(lowerURL, "caiyun.com"):
		return ProviderCaiyun, nil
	case strings.Contains(lowerURL, "cloud.189.cn") || strings.Contains(lowerURL, "tianyi.com"):
		return ProviderTianyi, nil
	case strings.Contains(lowerURL, "drive.uc.cn") || strings.Contains(lowerURL, "pan.uc.cn"):
		return ProviderUC, nil
	case strings.Contains(lowerURL, "pan.wo.cn") || strings.Contains(lowerURL, "wopan.cn"):
		return ProviderWopan, nil
	default:
		return "", ErrUnsupportedProvider
	}
}

// parseAliyunShareURL 解析阿里云盘分享URL
func parseAliyunShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://www.aliyundrive.com/s/xxxxx
	// https://www.alipan.com/s/xxxxx
	// https://aliyundrive.com/s/xxxxx
	// xxxx (直接分享ID)

	// 如果是纯ID（不含://），直接返回
	if !strings.Contains(urlStr, "://") {
		// 验证ID格式（字母数字，通常32位左右）
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	// 检查域名
	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "aliyundrive.com") &&
		!strings.Contains(host, "alipan.com") &&
		!strings.Contains(host, "aliyun.com") {
		return "", "", fmt.Errorf("%w: not a valid aliyun drive URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	// 尝试从查询参数中提取
	shareID = parsedURL.Query().Get("share_id")
	if shareID != "" {
		pwd = parsedURL.Query().Get("pwd")
		return shareID, pwd, nil
	}

	return "", "", ErrInvalidShareURL
}

// parseBaiduShareURL 解析百度网盘分享URL
func parseBaiduShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://pan.baidu.com/s/xxxxx
	// https://pan.baidu.com/share/init?surl=xxxxx
	// https://yun.baidu.com/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		// 纯分享ID
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "pan.baidu.com") && !strings.Contains(host, "yun.baidu.com") {
		return "", "", fmt.Errorf("%w: not a valid baidu pan URL", ErrInvalidShareURL)
	}

	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID = parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	// 尝试从 surl 参数提取
	surl := parsedURL.Query().Get("surl")
	if surl != "" {
		pwd = parsedURL.Query().Get("pwd")
		return surl, pwd, nil
	}

	return "", "", ErrInvalidShareURL
}

// parseXunleiShareURL 解析迅雷网盘分享URL
func parseXunleiShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://pan.xunlei.com/s/xxxxx
	// https://xunlei.com/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "xunlei.com") {
		return "", "", fmt.Errorf("%w: not a valid xunlei URL", ErrInvalidShareURL)
	}

	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID = parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parse115ShareURL 解析115网盘分享URL
func parse115ShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://115.com/s/xxxxx
	// https://115cdn.com/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "115.com") && !strings.Contains(host, "115cdn.com") {
		return "", "", fmt.Errorf("%w: not a valid 115 URL", ErrInvalidShareURL)
	}

	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "s" && i+1 < len(parts) {
			shareID = parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parseQuarkShareURL 解析夸克网盘分享URL
func parseQuarkShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://pan.quark.cn/s/xxxxx
	// https://quark.cn/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "quark.cn") {
		return "", "", fmt.Errorf("%w: not a valid quark URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parseCaiyunShareURL 解析和彩云分享URL
func parseCaiyunShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://caiyun.139.com/s/xxxxx
	// https://caiyun.139.com/w/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "caiyun.139.com") && !strings.Contains(host, "caiyun.com") {
		return "", "", fmt.Errorf("%w: not a valid caiyun URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parseTianyiShareURL 解析天翼云盘分享URL
func parseTianyiShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://cloud.189.cn/s/xxxxx
	// https://cloud.189.cn/t/xxxxx
	// https://cloud.189.cn/web/share?code=xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "cloud.189.cn") && !strings.Contains(host, "tianyi.com") {
		return "", "", fmt.Errorf("%w: not a valid tianyi URL", ErrInvalidShareURL)
	}

	// 检查是否是 /web/share?code=xxx 格式
	if parsedURL.Path == "/web/share" {
		code := parsedURL.Query().Get("code")
		if code != "" {
			return code, "", nil
		}
	}

	// 检查是否是 /s/xxx 或 /t/xxx 格式
	path := parsedURL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if (part == "s" || part == "t") && i+1 < len(parts) {
			shareID := parts[i+1]
			if idx := strings.Index(shareID, "?"); idx > 0 {
				shareID = shareID[:idx]
			}
			if shareID != "" {
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parseUCShareURL 解析UC网盘分享URL
func parseUCShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://drive.uc.cn/s/xxxxx
	// https://pan.uc.cn/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "drive.uc.cn") && !strings.Contains(host, "pan.uc.cn") {
		return "", "", fmt.Errorf("%w: not a valid UC URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parseWopanShareURL 解析联通云盘分享URL
func parseWopanShareURL(urlStr string) (shareID, pwd string, err error) {
	// 支持的URL格式:
	// https://pan.wo.cn/s/xxxxx
	// https://wopan.cn/s/xxxxx

	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "pan.wo.cn") && !strings.Contains(host, "wopan.cn") {
		return "", "", fmt.Errorf("%w: not a valid wopan URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}

// parsePikPakShareURL 解析PikPak分享URL
func parsePikPakShareURL(urlStr string) (shareID, pwd string, err error) {
	if !strings.Contains(urlStr, "://") {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, urlStr)
		if matched && len(urlStr) >= 10 {
			return urlStr, "", nil
		}
		return "", "", ErrInvalidShareURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", ErrInvalidShareURL
	}

	host := strings.ToLower(parsedURL.Host)
	if !strings.Contains(host, "mypikpak.com") && !strings.Contains(host, "pikpak.com") {
		return "", "", fmt.Errorf("%w: not a valid pikpak URL", ErrInvalidShareURL)
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
				pwd = parsedURL.Query().Get("pwd")
				return shareID, pwd, nil
			}
		}
	}

	return "", "", ErrInvalidShareURL
}
