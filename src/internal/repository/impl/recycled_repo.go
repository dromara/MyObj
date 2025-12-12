package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"time"

	"gorm.io/gorm"
)

type recycledRepository struct {
	db *gorm.DB
}

// NewRecycledRepository 创建回收站仓储实例
func NewRecycledRepository(db *gorm.DB) repository.RecycledRepository {
	return &recycledRepository{db: db}
}

func (r *recycledRepository) Create(ctx context.Context, recycled *models.Recycled) error {
	return r.db.WithContext(ctx).Create(recycled).Error
}

func (r *recycledRepository) GetByID(ctx context.Context, id string) (*models.Recycled, error) {
	var recycled models.Recycled
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&recycled).Error
	if err != nil {
		return nil, err
	}
	return &recycled, nil
}

func (r *recycledRepository) GetByUserIDAndFileID(ctx context.Context, userID, fileID string) (*models.Recycled, error) {
	var recycled models.Recycled
	err := r.db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, fileID).First(&recycled).Error
	if err != nil {
		return nil, err
	}
	return &recycled, nil
}

func (r *recycledRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Recycled{}).Error
}

func (r *recycledRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.Recycled, error) {
	var recycleds []*models.Recycled
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(limit).Find(&recycleds).Error
	return recycleds, err
}

func (r *recycledRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Recycled{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetExpiredRecords 获取超过指定天数的回收站记录
func (r *recycledRepository) GetExpiredRecords(ctx context.Context, days int) ([]*models.Recycled, error) {
	var recycleds []*models.Recycled
	expireTime := time.Now().AddDate(0, 0, -days)
	err := r.db.WithContext(ctx).
		Where("created_at < ?", expireTime).
		Find(&recycleds).Error
	return recycleds, err
}

// CountFileReferences 统计指定文件被多少个用户持有
func (r *recycledRepository) CountFileReferences(ctx context.Context, fileID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserFiles{}).
		Where("file_id = ?", fileID).Count(&count).Error
	return count, err
}
