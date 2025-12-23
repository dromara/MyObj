package service

import (
	"context"
	"errors"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecycledService 回收站服务
type RecycledService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewRecycledService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *RecycledService {
	return &RecycledService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}

func (r *RecycledService) GetRepository() *impl.RepositoryFactory {
	return r.factory
}

// GetRecycledList 获取回收站列表
func (r *RecycledService) GetRecycledList(req *request.RecycledListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	offset := (req.Page - 1) * req.PageSize

	// 查询回收站记录
	recycleds, err := r.factory.Recycled().ListByUserID(ctx, userID, offset, req.PageSize)
	if err != nil {
		logger.LOG.Error("查询回收站列表失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("查询回收站列表失败: %w", err)
	}

	// 统计总数
	total, err := r.factory.Recycled().Count(ctx, userID)
	if err != nil {
		logger.LOG.Error("统计回收站数量失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("统计回收站数量失败: %w", err)
	}

	// 构造响应数据
	items := make([]*response.RecycledItem, 0, len(recycleds))
	for _, recycled := range recycleds {
		// 获取用户文件关联，以获取文件名（使用 Unscoped 查询软删除的记录）
		var userFile models.UserFiles
		err = r.factory.DB().Unscoped().Where("user_id = ? AND uf_id = ?", userID, recycled.FileID).First(&userFile).Error
		if err != nil {
			logger.LOG.Warn("获取用户文件关联失败", "error", err, "userID", userID, "fileID", recycled.FileID)
			continue
		}
		fileInfo, err := r.factory.FileInfo().GetByID(ctx, userFile.FileID)
		if err != nil {
			logger.LOG.Warn("获取文件信息失败", "error", err, "userID", userID, "fileID", recycled.FileID)
			continue
		}
		items = append(items, &response.RecycledItem{
			RecycledID:   recycled.ID,
			FileID:       recycled.FileID,
			FileName:     userFile.FileName,
			FileSize:     int64(fileInfo.Size),
			MimeType:     fileInfo.Mime,
			IsEnc:        fileInfo.IsEnc,
			HasThumbnail: fileInfo.ThumbnailImg != "",
			DeletedAt:    recycled.CreatedAt,
		})
	}

	result := &response.RecycledListResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return models.NewJsonResponse(200, "获取成功", result), nil
}

// RestoreFile 还原文件
func (r *RecycledService) RestoreFile(req *request.RestoreFileRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证回收站记录是否存在且属于该用户
	recycled, err := r.factory.Recycled().GetByID(ctx, req.RecycledID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("回收站记录不存在")
		}
		logger.LOG.Error("获取回收站记录失败", "error", err, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("获取回收站记录失败: %w", err)
	}

	if recycled.UserID != userID {
		logger.LOG.Warn("用户尝试还原他人文件", "userID", userID, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("无权操作此文件")
	}

	// 获取要还原的文件记录（使用 Unscoped 查询软删除的记录）
	var userFile models.UserFiles
	err = r.factory.DB().Unscoped().Where("user_id = ? AND uf_id = ?", userID, recycled.FileID).First(&userFile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("文件记录不存在")
		}
		logger.LOG.Error("获取文件记录失败", "error", err, "userID", userID, "fileID", recycled.FileID)
		return nil, fmt.Errorf("获取文件记录失败: %w", err)
	}

	// 检查父目录是否存在
	var targetVirtualPath string = userFile.VirtualPath
	parentDirExists := false
	
	// 如果 VirtualPath 为空或 "0"，说明文件原本就在根目录，不需要检查
	if userFile.VirtualPath == "" || userFile.VirtualPath == "0" {
		parentDirExists = true // 根目录总是存在的
	} else {
		// 解析虚拟路径ID
		pathID := 0
		_, err := fmt.Sscanf(userFile.VirtualPath, "%d", &pathID)
		if err == nil && pathID > 0 {
			// 检查目录是否存在
			_, err := r.factory.VirtualPath().GetByID(ctx, pathID)
			if err == nil {
				parentDirExists = true
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 父目录不存在
				logger.LOG.Warn("文件原父目录已删除，将还原到根目录", 
					"userID", userID, 
					"fileID", recycled.FileID, 
					"originalPath", userFile.VirtualPath)
			} else {
				logger.LOG.Warn("检查父目录时出错", "error", err, "pathID", pathID)
			}
		}
	}

	// 如果父目录不存在，获取根目录ID
	if !parentDirExists {
		rootPath, err := r.factory.VirtualPath().GetRootPath(ctx, userID)
		if err != nil {
			logger.LOG.Error("获取根目录失败", "error", err, "userID", userID)
			return nil, fmt.Errorf("获取根目录失败: %w", err)
		}
		targetVirtualPath = fmt.Sprintf("%d", rootPath.ID)
		logger.LOG.Info("文件将还原到根目录", 
			"userID", userID, 
			"fileID", recycled.FileID, 
			"originalPath", userFile.VirtualPath, 
			"newPath", targetVirtualPath)
	}

	// 在事务中执行：1. 恢复 user_files 软删除、2. 更新 VirtualPath（如果父目录不存在）、3. 删除回收站记录
	err = r.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := r.factory.WithTx(tx)

		// 恢复 user_files 软删除（清除 deleted_at）
		// 如果父目录不存在，同时更新 VirtualPath
		updateMap := map[string]interface{}{
			"deleted_at": nil,
		}
		if !parentDirExists {
			updateMap["virtual_path"] = targetVirtualPath
		}
		
		if err := tx.Model(&models.UserFiles{}).Unscoped().
			Where("user_id = ? AND uf_id = ?", userID, recycled.FileID).
			Updates(updateMap).Error; err != nil {
			return fmt.Errorf("恢复用户文件失败: %w", err)
		}

		// 删除回收站记录
		if err := txFactory.Recycled().Delete(ctx, req.RecycledID); err != nil {
			return fmt.Errorf("删除回收站记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.LOG.Error("还原文件失败", "error", err, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("还原文件失败: %w", err)
	}

	message := "文件已还原"
	if !parentDirExists {
		message = "文件已还原到根目录（原父目录已删除）"
	}

	logger.LOG.Info("文件已还原", 
		"recycledID", req.RecycledID, 
		"userID", userID, 
		"fileID", recycled.FileID,
		"originalPath", userFile.VirtualPath,
		"newPath", targetVirtualPath)
	return models.NewJsonResponse(200, message, nil), nil
}

// DeletePermanently 永久删除文件
func (r *RecycledService) DeletePermanently(req *request.DeleteRecycledRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证回收站记录
	recycled, err := r.factory.Recycled().GetByID(ctx, req.RecycledID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("回收站记录不存在")
		}
		logger.LOG.Error("获取回收站记录失败", "error", err, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("获取回收站记录失败: %w", err)
	}

	if recycled.UserID != userID {
		logger.LOG.Warn("用户尝试删除他人文件", "userID", userID, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("无权操作此文件")
	}

	// 执行永久删除
	if err := r.deleteSingleFile(ctx, recycled); err != nil {
		logger.LOG.Error("永久删除文件失败", "error", err, "recycledID", req.RecycledID)
		return nil, fmt.Errorf("永久删除文件失败: %w", err)
	}

	logger.LOG.Info("文件已永久删除", "recycledID", req.RecycledID, "userID", userID, "fileID", recycled.FileID)
	return models.NewJsonResponse(200, "文件已永久删除", nil), nil
}

// EmptyRecycled 清空回收站
func (r *RecycledService) EmptyRecycled(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 获取该用户的所有回收站记录
	recycleds, err := r.factory.Recycled().ListByUserID(ctx, userID, 0, 10000) // 每次清除10000个文件
	if err != nil {
		logger.LOG.Error("查询回收站列表失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("查询回收站列表失败: %w", err)
	}

	deletedCount := 0
	failedCount := 0

	// 逐个删除
	for _, recycled := range recycleds {
		if err := r.deleteSingleFile(ctx, recycled); err != nil {
			logger.LOG.Error("删除文件失败", "error", err, "recycledID", recycled.ID)
			failedCount++
		} else {
			deletedCount++
		}
	}

	logger.LOG.Info("清空回收站完成",
		"userID", userID,
		"deleted", deletedCount,
		"failed", failedCount)

	message := fmt.Sprintf("已清空回收站，成功删除 %d 个文件", deletedCount)
	if failedCount > 0 {
		message = fmt.Sprintf("%s，失败 %d 个", message, failedCount)
	}

	return models.NewJsonResponse(200, message, map[string]int{
		"deleted": deletedCount,
		"failed":  failedCount,
	}), nil
}

// MoveToRecycled 将文件移动到回收站
func (r *RecycledService) MoveToRecycled(fileID, userID string) error {
	ctx := context.Background()

	// 创建回收站记录
	recycled := &models.Recycled{
		ID:        uuid.Must(uuid.NewV7()).String(),
		FileID:    fileID,
		UserID:    userID,
		CreatedAt: custom_type.Now(),
	}

	if err := r.factory.Recycled().Create(ctx, recycled); err != nil {
		logger.LOG.Error("创建回收站记录失败", "error", err, "fileID", fileID, "userID", userID)
		return fmt.Errorf("移动到回收站失败: %w", err)
	}

	logger.LOG.Info("文件已移动到回收站", "fileID", fileID, "userID", userID)
	return nil
}

// deleteSingleFile 删除单个文件（参考定时任务的逻辑）
func (r *RecycledService) deleteSingleFile(ctx context.Context, recycled *models.Recycled) error {
	// 1. 检查文件引用数
	var userFile *models.UserFiles
	err := r.factory.DB().Unscoped().Where("user_id = ? AND uf_id = ?", recycled.UserID, recycled.FileID).First(&userFile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LOG.Warn("用户文件记录不存在，直接删除回收站记录", "file_id", recycled.FileID)
			return r.factory.Recycled().Delete(ctx, recycled.ID)
		}
		logger.LOG.Error("获取用户文件记录失败", "error", err, "file_id", recycled.FileID)
		return err
	}
	refCount, err := r.factory.Recycled().CountFileReferences(ctx, userFile.FileID)
	if err != nil {
		return fmt.Errorf("统计文件引用失败: %w", err)
	}

	// 2. 如果引用数 > 1，说明其他用户也持有该文件，仅删除回收站记录
	if refCount > 1 {
		logger.LOG.Debug("文件被多个用户持有，仅删除回收站记录",
			"file_id", recycled.FileID,
			"ref_count", refCount)
		return r.factory.Recycled().Delete(ctx, recycled.ID)
	}

	// 3. 获取文件信息
	fileInfo, err := r.factory.FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LOG.Warn("文件信息不存在，直接删除回收站记录", "file_id", recycled.FileID)
			return r.factory.Recycled().Delete(ctx, recycled.ID)
		}
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 4. 获取用户信息（用于空间归还）
	user, err := r.factory.User().GetByID(ctx, recycled.UserID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 5. 在事务中执行删除操作
	return r.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := r.factory.WithTx(tx)

		// 5.1 删除物理文件（普通文件或加密文件）
		if err := r.deletePhysicalFile(fileInfo); err != nil {
			logger.LOG.Warn("删除物理文件失败", "error", err)
			// 物理文件删除失败不阻塞事务，继续执行
		}

		// 5.2 删除缩略图
		if fileInfo.ThumbnailImg != "" {
			if err := r.deleteThumbnail(fileInfo.ThumbnailImg); err != nil {
				logger.LOG.Warn("删除缩略图失败", "error", err)
			}
		}

		// 5.3 删除用户文件关联
		if err := txFactory.UserFiles().Delete(ctx, recycled.UserID, recycled.FileID); err != nil {
			return fmt.Errorf("删除用户文件关联失败: %w", err)
		}

		// 5.4 如果是分片文件，删除所有分片记录
		if fileInfo.IsChunk {
			if err := txFactory.FileChunk().DeleteByFileID(ctx, recycled.FileID); err != nil {
				return fmt.Errorf("删除文件分片记录失败: %w", err)
			}
		}

		// 5.5 删除FileInfo记录
		if err := txFactory.FileInfo().Delete(ctx, userFile.FileID); err != nil {
			return fmt.Errorf("删除文件信息记录失败: %w", err)
		}

		// 5.6 删除回收站记录
		if err := txFactory.Recycled().Delete(ctx, recycled.ID); err != nil {
			return fmt.Errorf("删除回收站记录失败: %w", err)
		}

		// 5.7 归还用户空间（只对非无限空间用户）
		if user.Space > 0 {
			user.FreeSpace += int64(fileInfo.Size)
			if err := txFactory.User().Update(ctx, user); err != nil {
				return fmt.Errorf("更新用户空间失败: %w", err)
			}
			logger.LOG.Debug("归还用户空间",
				"user_id", user.ID,
				"returned_size", fileInfo.Size,
				"new_free_space", user.FreeSpace)
		}
		return nil
	})
}

// deletePhysicalFile 删除物理文件
func (r *RecycledService) deletePhysicalFile(fileInfo *models.FileInfo) error {
	// 如果有加密文件，优先删除加密文件
	if fileInfo.IsEnc && fileInfo.EncPath != "" {
		if err := r.deleteFile(fileInfo.EncPath); err != nil {
			logger.LOG.Warn("删除加密文件失败", "path", fileInfo.EncPath, "error", err)
		}
		// 删除.info文件
		infoPath := fileInfo.EncPath + ".info"
		if err := r.deleteFile(infoPath); err != nil {
			logger.LOG.Warn("删除.info文件失败", "path", infoPath, "error", err)
		}
	}

	// 删除普通文件
	if fileInfo.Path != "" {
		if err := r.deleteFile(fileInfo.Path); err != nil {
			logger.LOG.Warn("删除普通文件失败", "path", fileInfo.Path, "error", err)
		}
		ext := filepath.Ext(fileInfo.Path)
		basePath := strings.TrimSuffix(fileInfo.Path, ext)
		infoPath := basePath + ".info"
		if err := r.deleteFile(infoPath); err != nil {
			logger.LOG.Warn("删除.info文件失败", "path", infoPath, "error", err)
		}

	}

	// 如果是分片文件，删除分片目录
	if fileInfo.IsChunk && fileInfo.Path != "" {
		// 文件路径格式: {DataPath}/data/{\u539f文件名不带后缀}/{\u865a拟文件名}.data
		// 分片目录为: {DataPath}/data/{\u539f文件名不带后缀}/{\u865a拟文件名}
		chunkDir := strings.TrimSuffix(fileInfo.Path, ".data")
		if err := r.deleteDirectory(chunkDir); err != nil {
			logger.LOG.Warn("删除分片目录失败", "path", chunkDir, "error", err)
		}
		// 删除父目录（如果为空）
		// 路径格式: {DataPath}/data/{\u539f文件名不带后缀}
		parentDir := filepath.Dir(fileInfo.Path)
		if err := r.deleteDirectoryIfEmpty(parentDir); err != nil {
			logger.LOG.Warn("删除父目录失败", "path", parentDir, "error", err)
		}
	} else if fileInfo.Path != "" {
		// 对于非分片文件，删除 .data 文件所在的文件夹（如果为空）
		// 路径格式: {DataPath}/data/{\u539f文件名不带后缀}/{\u865a拟文件名}.data
		parentDir := filepath.Dir(fileInfo.Path)
		if err := r.deleteDirectoryIfEmpty(parentDir); err != nil {
			logger.LOG.Warn("删除文件夹失败", "path", parentDir, "error", err)
		}
	}

	return nil
}

// deleteThumbnail 删除缩略图
func (r *RecycledService) deleteThumbnail(thumbnailPath string) error {
	return r.deleteFile(thumbnailPath)
}

// deleteFile 删除文件
func (r *RecycledService) deleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.LOG.Debug("文件不存在，跳过删除", "path", filePath)
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %w", filePath, err)
	}

	logger.LOG.Debug("成功删除文件", "path", filePath)
	return nil
}

// deleteDirectory 删除目录
func (r *RecycledService) deleteDirectory(dirPath string) error {
	if dirPath == "" {
		return nil
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logger.LOG.Debug("目录不存在，跳过删除", "path", dirPath)
		return nil
	}

	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("删除目录失败 %s: %w", dirPath, err)
	}

	logger.LOG.Debug("成功删除目录", "path", dirPath)
	return nil
}

// deleteDirectoryIfEmpty 删除空目录（如果目录为空）
func (r *RecycledService) deleteDirectoryIfEmpty(dirPath string) error {
	if dirPath == "" {
		return nil
	}

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		logger.LOG.Debug("目录不存在，跳过删除", "path", dirPath)
		return nil
	}

	// 读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取目录失败 %s: %w", dirPath, err)
	}

	// 如果目录不为空，不删除
	if len(entries) > 0 {
		logger.LOG.Debug("目录不为空，跳过删除", "path", dirPath, "file_count", len(entries))
		return nil
	}

	// 删除空目录
	if err := os.Remove(dirPath); err != nil {
		return fmt.Errorf("删除空目录失败 %s: %w", dirPath, err)
	}

	logger.LOG.Debug("成功删除空目录", "path", dirPath)
	return nil
}
