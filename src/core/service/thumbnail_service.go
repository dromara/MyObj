package service

import (
	"context"
	"fmt"
	"myobj/src/config"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/preview"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
)

// ThumbnailService 缩略图服务
type ThumbnailService struct {
	factory *impl.RepositoryFactory
}

// NewThumbnailService 创建缩略图服务实例
func NewThumbnailService(factory *impl.RepositoryFactory) *ThumbnailService {
	return &ThumbnailService{factory: factory}
}

// GenerateVideoThumbnail 使用ffmpeg生成视频缩略图
//
// 参数:
//   - ctx: 上下文
//   - fileID: file_info 表的文件ID
//   - force: 是否强制重新生成（覆盖已有缩略图）
//
// 返回:
//   - thumbnailPath: 生成的缩略图路径
//   - err: 错误信息
func (s *ThumbnailService) GenerateVideoThumbnail(ctx context.Context, fileID string, force bool) (string, error) {
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	if !preview.IsVideoByMime(fileInfo.Mime) {
		return "", fmt.Errorf("文件不是视频类型: %s", fileInfo.Mime)
	}

	return s.generateThumbnail(ctx, fileInfo, force, func(outputPath string, maxDimension uint) error {
		return preview.GenerateVideoThumbnail(fileInfo.Path, outputPath, "", maxDimension)
	})
}

// GenerateImageThumbnail 生成图片缩略图
//
// 参数:
//   - ctx: 上下文
//   - fileID: file_info 表的文件ID
//   - force: 是否强制重新生成（覆盖已有缩略图）
//
// 返回:
//   - thumbnailPath: 生成的缩略图路径
//   - err: 错误信息
func (s *ThumbnailService) GenerateImageThumbnail(ctx context.Context, fileID string, force bool) (string, error) {
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	if !util.IsImageByMime(fileInfo.Mime) {
		return "", fmt.Errorf("文件不是图片类型: %s", fileInfo.Mime)
	}

	return s.generateThumbnail(ctx, fileInfo, force, func(outputPath string, maxDimension uint) error {
		return preview.GenerateImageThumbnail(fileInfo.Path, outputPath, maxDimension)
	})
}

// GetThumbnail 获取缩略图（不存在则自动生成）
//
// 参数:
//   - ctx: 上下文
//   - fileID: file_info 表的文件ID
//
// 返回:
//   - thumbnailPath: 缩略图路径
//   - err: 错误信息
func (s *ThumbnailService) GetThumbnail(ctx context.Context, fileID string) (string, error) {
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	if fileInfo.IsEnc {
		return "", fmt.Errorf("加密文件无法生成缩略图")
	}

	if fileInfo.ThumbnailImg != "" {
		if _, err := os.Stat(fileInfo.ThumbnailImg); err == nil {
			return fileInfo.ThumbnailImg, nil
		}
	}

	if !config.CONFIG.File.Thumbnail {
		return "", fmt.Errorf("缩略图功能未启用")
	}

	var generateFn func(outputPath string, maxDimension uint) error
	if preview.IsVideoByMime(fileInfo.Mime) {
		if !preview.CheckFFmpegAvailable() {
			return "", fmt.Errorf("ffmpeg 不可用，无法生成视频缩略图")
		}
		generateFn = func(outputPath string, maxDimension uint) error {
			return preview.GenerateVideoThumbnail(fileInfo.Path, outputPath, "", maxDimension)
		}
	} else if util.IsImageByMime(fileInfo.Mime) {
		generateFn = func(outputPath string, maxDimension uint) error {
			return preview.GenerateImageThumbnail(fileInfo.Path, outputPath, maxDimension)
		}
	} else {
		return "", fmt.Errorf("不支持的文件类型: %s", fileInfo.Mime)
	}

	return s.generateThumbnail(ctx, fileInfo, true, generateFn)
}

// generateThumbnail 通用缩略图生成逻辑
func (s *ThumbnailService) generateThumbnail(ctx context.Context, fileInfo *models.FileInfo, force bool, generateFn func(string, uint) error) (string, error) {
	if fileInfo.IsEnc {
		return "", fmt.Errorf("加密文件无法生成缩略图")
	}

	if !force && fileInfo.ThumbnailImg != "" {
		if _, err := os.Stat(fileInfo.ThumbnailImg); err == nil {
			return fileInfo.ThumbnailImg, nil
		}
	}

	if !config.CONFIG.File.Thumbnail {
		return "", fmt.Errorf("缩略图功能未启用")
	}

	maxDimension := uint(300)

	thumbnailPath, err := s.buildThumbnailPath(ctx, fileInfo)
	if err != nil {
		return "", err
	}

	if err := generateFn(thumbnailPath, maxDimension); err != nil {
		return "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	fileInfo.ThumbnailImg = thumbnailPath
	fileInfo.UpdatedAt = custom_type.Now()
	if err := s.factory.FileInfo().Update(ctx, fileInfo); err != nil {
		logger.LOG.Warn("更新文件缩略图路径失败", "error", err, "fileID", fileInfo.ID)
	}

	return thumbnailPath, nil
}

// buildThumbnailPath 构建缩略图存储路径: data/thumbnails/{file_id}_thumb.jpg
func (s *ThumbnailService) buildThumbnailPath(ctx context.Context, fileInfo *models.FileInfo) (string, error) {
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		return "", fmt.Errorf("查询磁盘列表失败: %w", err)
	}
	if len(disks) == 0 {
		return "", fmt.Errorf("没有可用的存储磁盘")
	}

	var bestDisk *models.Disk
	var maxFree int64
	for _, disk := range disks {
		freeSpace, err := util.GetDiskFreeSpaceByPath(disk.DataPath)
		if err != nil {
			continue
		}
		if freeSpace > maxFree {
			maxFree = freeSpace
			bestDisk = disk
		}
	}
	if bestDisk == nil {
		bestDisk = disks[0]
	}

	thumbnailDir := filepath.Join(bestDisk.DataPath, "thumbnails")
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return "", fmt.Errorf("创建缩略图目录失败: %w", err)
	}

	return filepath.Join(thumbnailDir, fileInfo.ID+"_thumb.jpg"), nil
}

// GenerateForFile 为指定文件生成缩略图（支持图片和视频）
//
// 参数:
//   - ctx: 上下文
//   - fileID: file_info 表的文件ID
//   - force: 是否强制重新生成（覆盖已有缩略图）
//
// 返回:
//   - thumbnailPath: 生成的缩略图路径
//   - err: 错误信息
func (s *ThumbnailService) GenerateForFile(ctx context.Context, fileID string, force bool) (string, error) {
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	if fileInfo.IsEnc {
		return "", fmt.Errorf("加密文件无法生成缩略图")
	}

	if !force && fileInfo.ThumbnailImg != "" {
		if _, err := os.Stat(fileInfo.ThumbnailImg); err == nil {
			return fileInfo.ThumbnailImg, nil
		}
	}

	if !config.CONFIG.File.Thumbnail {
		return "", fmt.Errorf("缩略图功能未启用")
	}

	var thumbnailPath string
	maxDimension := uint(300)

	if preview.IsVideoByMime(fileInfo.Mime) {
		thumbnailPath, err = s.generateVideoThumbnail(fileInfo, maxDimension)
	} else if util.IsImageByMime(fileInfo.Mime) {
		thumbnailPath, err = s.generateImageThumbnail(fileInfo, maxDimension)
	} else {
		return "", fmt.Errorf("不支持的文件类型: %s", fileInfo.Mime)
	}

	if err != nil {
		return "", err
	}

	fileInfo.ThumbnailImg = thumbnailPath
	fileInfo.UpdatedAt = custom_type.Now()
	if err := s.factory.FileInfo().Update(ctx, fileInfo); err != nil {
		logger.LOG.Warn("更新文件缩略图路径失败", "error", err, "fileID", fileID)
	}

	return thumbnailPath, nil
}

// generateVideoThumbnail 生成视频缩略图
func (s *ThumbnailService) generateVideoThumbnail(fileInfo *models.FileInfo, maxDimension uint) (string, error) {
	if !preview.CheckFFmpegAvailable() {
		return "", fmt.Errorf("ffmpeg 不可用，无法生成视频缩略图")
	}

	storageDir := filepath.Dir(fileInfo.Path)
	thumbnailPath := filepath.Join(storageDir, fileInfo.RandomName+".jpg")

	if err := preview.GenerateVideoThumbnail(fileInfo.Path, thumbnailPath, "", maxDimension); err != nil {
		return "", fmt.Errorf("生成视频缩略图失败: %w", err)
	}

	return thumbnailPath, nil
}

// generateImageThumbnail 生成图片缩略图
func (s *ThumbnailService) generateImageThumbnail(fileInfo *models.FileInfo, maxDimension uint) (string, error) {
	storageDir := filepath.Dir(fileInfo.Path)
	thumbnailPath := filepath.Join(storageDir, fileInfo.RandomName+".jpg")

	if err := preview.GenerateImageThumbnail(fileInfo.Path, thumbnailPath, maxDimension); err != nil {
		return "", fmt.Errorf("生成图片缩略图失败: %w", err)
	}

	return thumbnailPath, nil
}

// GenerateVideoThumbnailAsync 异步生成视频缩略图（不阻塞主流程）
// 返回一个 channel，生成完成后会发送结果
func (s *ThumbnailService) GenerateVideoThumbnailAsync(fileInfo *models.FileInfo, maxDimension uint) <-chan struct {
	Path string
	Err  error
} {
	resultChan := make(chan struct {
		Path string
		Err  error
	}, 1)

	go func() {
		defer close(resultChan)
		path, err := s.generateVideoThumbnail(fileInfo, maxDimension)
		resultChan <- struct {
			Path string
			Err  error
		}{Path: path, Err: err}
	}()

	return resultChan
}
