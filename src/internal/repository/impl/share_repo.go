package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type shareRepository struct {
	db *gorm.DB
}

// NewShareRepository 创建分享仓储实例
func NewShareRepository(db *gorm.DB) repository.ShareRepository {
	return &shareRepository{db: db}
}

func (r *shareRepository) Create(ctx context.Context, share *models.Share) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *shareRepository) GetByID(ctx context.Context, id int) (*models.Share, error) {
	var share models.Share
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) GetByToken(ctx context.Context, token string) (*models.Share, error) {
	var share models.Share
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *shareRepository) Update(ctx context.Context, share *models.Share) error {
	return r.db.WithContext(ctx).Save(share).Error
}

func (r *shareRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Share{}).Error
}

func (r *shareRepository) List(ctx context.Context, userID string, offset, limit int) ([]*models.Share, error) {
	var shares []*models.Share
	query := r.db.WithContext(ctx).Offset(offset).Limit(limit)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Find(&shares).Error
	return shares, err
}

func (r *shareRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Share{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *shareRepository) IncrementDownloadCount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&models.Share{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error
}
