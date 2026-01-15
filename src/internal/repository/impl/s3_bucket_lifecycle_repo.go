package impl

import (
	"context"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"gorm.io/gorm"
)

// S3BucketLifecycleRepositoryImpl S3 Bucket Lifecycle配置仓储实现
type S3BucketLifecycleRepositoryImpl struct {
	db *gorm.DB
}

// NewS3BucketLifecycleRepository 创建S3 Bucket Lifecycle配置仓储实例
func NewS3BucketLifecycleRepository(db *gorm.DB) *S3BucketLifecycleRepositoryImpl {
	return &S3BucketLifecycleRepositoryImpl{db: db}
}

// GetByBucket 根据Bucket名称获取Lifecycle配置
func (r *S3BucketLifecycleRepositoryImpl) GetByBucket(ctx context.Context, bucketName, userID string) (*models.S3BucketLifecycle, error) {
	var lifecycle models.S3BucketLifecycle
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		First(&lifecycle).Error
	if err != nil {
		return nil, err
	}
	return &lifecycle, nil
}

// CreateOrUpdate 创建或更新Lifecycle配置
func (r *S3BucketLifecycleRepositoryImpl) CreateOrUpdate(ctx context.Context, lifecycle *models.S3BucketLifecycle) error {
	// 先尝试查找是否存在
	existing, err := r.GetByBucket(ctx, lifecycle.BucketName, lifecycle.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		lifecycle.ID = existing.ID
		lifecycle.CreatedAt = existing.CreatedAt
		if lifecycle.CreatedAt.IsZero() {
			lifecycle.CreatedAt = existing.CreatedAt
		}
		return r.db.WithContext(ctx).Save(lifecycle).Error
	}

	// 创建
	if lifecycle.CreatedAt.IsZero() {
		lifecycle.CreatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(lifecycle).Error
}

// Delete 删除Lifecycle配置
func (r *S3BucketLifecycleRepositoryImpl) Delete(ctx context.Context, bucketName, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		Delete(&models.S3BucketLifecycle{}).Error
}

// ListAll 列出所有Lifecycle配置（用于定时任务）
func (r *S3BucketLifecycleRepositoryImpl) ListAll(ctx context.Context) ([]*models.S3BucketLifecycle, error) {
	var lifecycles []*models.S3BucketLifecycle
	err := r.db.WithContext(ctx).Find(&lifecycles).Error
	return lifecycles, err
}
