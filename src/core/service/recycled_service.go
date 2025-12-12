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
		// 获取文件信息
		fileInfo, err := r.factory.FileInfo().GetByID(ctx, recycled.FileID)
		if err != nil {
			logger.LOG.Warn("获取文件信息失败", "error", err, "fileID", recycled.FileID)
			continue
		}

		// 获取用户文件关联，以获取文件名（使用 Unscoped 查询软删除的记录）
		var userFile models.UserFiles
		err = r.factory.DB().Unscoped().Where("user_id = ? AND file_id = ?", userID, recycled.FileID).First(&userFile).Error
		if err != nil {
			logger.LOG.Warn("获取用户文件关联失败", "error", err, "userID", userID, "fileID", recycled.FileID)
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

	// 在事务中执行：1. 恢复 user_files 软删除、 2. 删除回收站记录
	err = r.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := r.factory.WithTx(tx)

		// 恢复 user_files 软删除（清除 deleted_at）
		if err := tx.Model(&models.UserFiles{}).Unscoped().
			Where("user_id = ? AND file_id = ?", userID, recycled.FileID).
			Update("deleted_at", nil).Error; err != nil {
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

	logger.LOG.Info("文件已还原", "recycledID", req.RecycledID, "userID", userID, "fileID", recycled.FileID)
	return models.NewJsonResponse(200, "文件已还原", nil), nil
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
	recycleds, err := r.factory.Recycled().ListByUserID(ctx, userID, 0, 10000) // 假设最多10000条
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

	// 检查是否已在回收站
	_, err := r.factory.Recycled().GetByUserIDAndFileID(ctx, userID, fileID)
	if err == nil {
		return fmt.Errorf("文件已在回收站中")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询回收站失败: %w", err)
	}

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
	refCount, err := r.factory.Recycled().CountFileReferences(ctx, recycled.FileID)
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
	fileInfo, err := r.factory.FileInfo().GetByID(ctx, recycled.FileID)
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

		// 5.1 删除用户文件关联
		if err := txFactory.UserFiles().Delete(ctx, recycled.UserID, recycled.FileID); err != nil {
			return fmt.Errorf("删除用户文件关联失败: %w", err)
		}

		// 5.2 如果是分片文件，删除所有分片记录
		if fileInfo.IsChunk {
			if err := txFactory.FileChunk().DeleteByFileID(ctx, recycled.FileID); err != nil {
				return fmt.Errorf("删除文件分片记录失败: %w", err)
			}
		}

		// 5.3 删除FileInfo记录
		if err := txFactory.FileInfo().Delete(ctx, recycled.FileID); err != nil {
			return fmt.Errorf("删除文件信息记录失败: %w", err)
		}

		// 5.4 删除回收站记录
		if err := txFactory.Recycled().Delete(ctx, recycled.ID); err != nil {
			return fmt.Errorf("删除回收站记录失败: %w", err)
		}

		// 5.5 归还用户空间（只对非无限空间用户）
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
