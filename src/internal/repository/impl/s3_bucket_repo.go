package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

// S3BucketRepositoryImpl S3 Bucket仓储实现
type S3BucketRepositoryImpl struct {
	db *gorm.DB
}

// NewS3BucketRepository 创建S3 Bucket仓储实例
func NewS3BucketRepository(db *gorm.DB) repository.S3BucketRepository {
	return &S3BucketRepositoryImpl{db: db}
}

// Create 创建Bucket
func (r *S3BucketRepositoryImpl) Create(ctx context.Context, bucket *models.S3Bucket) error {
	return r.db.WithContext(ctx).Create(bucket).Error
}

// GetByName 根据Bucket名称获取
func (r *S3BucketRepositoryImpl) GetByName(ctx context.Context, bucketName string, userID string) (*models.S3Bucket, error) {
	var bucket models.S3Bucket
	err := r.db.WithContext(ctx).Where("bucket_name = ? AND user_id = ?", bucketName, userID).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// GetByID 根据ID获取
func (r *S3BucketRepositoryImpl) GetByID(ctx context.Context, id int) (*models.S3Bucket, error) {
	var bucket models.S3Bucket
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&bucket).Error
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

// ListByUserID 列出用户的所有Bucket
func (r *S3BucketRepositoryImpl) ListByUserID(ctx context.Context, userID string) ([]*models.S3Bucket, error) {
	var buckets []*models.S3Bucket
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&buckets).Error
	return buckets, err
}

// Update 更新Bucket
func (r *S3BucketRepositoryImpl) Update(ctx context.Context, bucket *models.S3Bucket) error {
	return r.db.WithContext(ctx).Save(bucket).Error
}

// Delete 删除Bucket
func (r *S3BucketRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.S3Bucket{}, id).Error
}

// Exists 检查Bucket是否存在
func (r *S3BucketRepositoryImpl) Exists(ctx context.Context, bucketName string, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.S3Bucket{}).Where("bucket_name = ? AND user_id = ?", bucketName, userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
