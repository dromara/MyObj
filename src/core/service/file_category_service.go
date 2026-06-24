package service

import (
	"context"
	"fmt"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"path/filepath"
	"strings"
	"time"
)

// FileCategory 文件分类
type FileCategory string

const (
	CategoryDocument FileCategory = "document"
	CategoryImage    FileCategory = "image"
	CategoryVideo    FileCategory = "video"
	CategoryAudio    FileCategory = "audio"
	CategoryArchive  FileCategory = "archive"
	CategoryCode     FileCategory = "code"
	CategoryOther    FileCategory = "other"
)

// categoryRules 分类规则（扩展名 -> 分类）
var categoryRules = map[FileCategory][]string{
	CategoryDocument: {".doc", ".docx", ".pdf", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".rtf", ".csv", ".odt", ".ods", ".odp"},
	CategoryImage:    {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff", ".tif", ".psd", ".raw"},
	CategoryVideo:    {".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp", ".ts"},
	CategoryAudio:    {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".opus", ".amr", ".mid", ".midi"},
	CategoryArchive:  {".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".zst", ".tgz", ".tbz2", ".cab", ".iso"},
	CategoryCode:     {".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".html", ".css", ".scss", ".less", ".json", ".xml", ".yaml", ".yml", ".toml", ".sh", ".bat", ".ps1", ".sql", ".rb", ".php", ".rs", ".swift", ".kt", ".vue", ".jsx", ".tsx"},
}

// extToCategory 扩展名到分类的反向索引（由 init 构建）
var extToCategory map[string]FileCategory

func init() {
	extToCategory = make(map[string]FileCategory)
	for cat, exts := range categoryRules {
		for _, ext := range exts {
			extToCategory[strings.ToLower(ext)] = cat
		}
	}
}

// mimeCategoryRules MIME类型前缀到分类的映射
var mimeCategoryRules = map[string]FileCategory{
	"image/":       CategoryImage,
	"video/":       CategoryVideo,
	"audio/":       CategoryAudio,
	"text/":        CategoryDocument,
	"application/pdf":    CategoryDocument,
	"application/msword": CategoryDocument,
	"application/vnd.openxmlformats": CategoryDocument,
	"application/vnd.ms-excel":       CategoryDocument,
	"application/vnd.ms-powerpoint":  CategoryDocument,
	"application/zip":                CategoryArchive,
	"application/x-rar":              CategoryArchive,
	"application/x-7z":               CategoryArchive,
	"application/x-tar":              CategoryArchive,
	"application/gzip":               CategoryArchive,
	"application/x-bzip2":            CategoryArchive,
}

// customCategoryRules 自定义分类规则（用户可扩展）
var customCategoryRules = make(map[string]FileCategory)

// FileCategoryService 文件分类服务
type FileCategoryService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewFileCategoryService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *FileCategoryService {
	return &FileCategoryService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}

func (s *FileCategoryService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// GetCategoryByExt 根据文件扩展名获取分类
func (s *FileCategoryService) GetCategoryByExt(ext string) FileCategory {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// 优先检查自定义规则
	if cat, ok := customCategoryRules[ext]; ok {
		return cat
	}

	// 检查内置规则
	if cat, ok := extToCategory[ext]; ok {
		return cat
	}

	return CategoryOther
}

// GetCategoryByMIME 根据MIME类型获取分类
func (s *FileCategoryService) GetCategoryByMIME(mime string) FileCategory {
	mime = strings.ToLower(mime)

	// 优先检查自定义规则
	if cat, ok := customCategoryRules[mime]; ok {
		return cat
	}

	// 精确匹配
	if cat, ok := mimeCategoryRules[mime]; ok {
		return cat
	}

	// 前缀匹配
	for prefix, cat := range mimeCategoryRules {
		if strings.HasSuffix(prefix, "/") {
			if strings.HasPrefix(mime, prefix) {
				return cat
			}
		}
	}

	return CategoryOther
}

// AutoClassify 自动分类文件（优先使用扩展名，回退到MIME类型）
func (s *FileCategoryService) AutoClassify(fileName string, mime string) FileCategory {
	ext := filepath.Ext(fileName)
	if ext != "" {
		cat := s.GetCategoryByExt(ext)
		if cat != CategoryOther {
			return cat
		}
	}

	if mime != "" {
		return s.GetCategoryByMIME(mime)
	}

	return CategoryOther
}

// AddCustomRule 添加自定义分类规则
// extOrMime: 扩展名（如 ".xyz"）或MIME类型（如 "application/custom"）
// category: 目标分类
func (s *FileCategoryService) AddCustomRule(extOrMime string, category FileCategory) {
	customCategoryRules[strings.ToLower(extOrMime)] = category
}

// GetAllCategories 获取所有可用分类
func (s *FileCategoryService) GetAllCategories() []FileCategory {
	return []FileCategory{
		CategoryDocument,
		CategoryImage,
		CategoryVideo,
		CategoryAudio,
		CategoryArchive,
		CategoryCode,
		CategoryOther,
	}
}

// GetUserCategoryStats 获取用户分类统计
func (s *FileCategoryService) GetUserCategoryStats(ctx context.Context, userID string) (*response.CategoryStatsResponse, error) {
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("category_stats:%s", userID)
	if cached, err := s.cacheLocal.Get(cacheKey); err == nil && cached != nil {
		if stats, ok := cached.(*response.CategoryStatsResponse); ok {
			return stats, nil
		}
	}

	stats, err := s.factory.FileInfo().CountByCategoryForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := &response.CategoryStatsResponse{
		Categories: make([]response.CategoryStat, 0, len(stats)),
	}

	for _, stat := range stats {
		result.Categories = append(result.Categories, response.CategoryStat{
			Category:  stat.Category,
			Count:     stat.Count,
			TotalSize: stat.TotalSize,
		})
		result.TotalCount += stat.Count
		result.TotalSize += stat.TotalSize
	}

	// 确保所有分类都有返回（即使为0）
	existingCats := make(map[string]bool)
	for _, cat := range result.Categories {
		existingCats[cat.Category] = true
	}
	for _, cat := range s.GetAllCategories() {
		if !existingCats[string(cat)] {
			result.Categories = append(result.Categories, response.CategoryStat{
				Category:  string(cat),
				Count:     0,
				TotalSize: 0,
			})
		}
	}

	// 缓存5分钟
	if err := s.cacheLocal.Set(cacheKey, result, 300); err != nil {
		logger.LOG.Warn("缓存分类统计失败", "error", err)
	}

	return result, nil
}

// InvalidateUserStatsCache 清除用户分类统计缓存
func (s *FileCategoryService) InvalidateUserStatsCache(userID string) {
	cacheKey := fmt.Sprintf("category_stats:%s", userID)
	if err := s.cacheLocal.Delete(cacheKey); err != nil {
		logger.LOG.Warn("清除分类统计缓存失败", "error", err)
	}
}

// CategoryStatsItem 分类统计查询结果
type CategoryStatsItem struct {
	Category  string
	Count     int64
	TotalSize int64
}

// GetAllCategoriesWithInfo 获取所有分类的详细信息（包含默认扩展名列表）
func (s *FileCategoryService) GetAllCategoriesWithInfo() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(categoryRules))
	for cat, exts := range categoryRules {
		result = append(result, map[string]interface{}{
			"category":   string(cat),
			"extensions": exts,
		})
	}
	return result
}

// WaitForCacheExpiry 等待缓存过期（用于测试）
func (s *FileCategoryService) WaitForCacheExpiry() {
	time.Sleep(310 * time.Second)
}
