package impl

import (
	"context"
	"myobj/src/pkg/models"

	"gorm.io/gorm"
)

// S3BucketCORSRepositoryImpl S3 Bucket CORS配置仓储实现
type S3BucketCORSRepositoryImpl struct {
	db *gorm.DB
}

// NewS3BucketCORSRepository 创建S3 Bucket CORS配置仓储实例
func NewS3BucketCORSRepository(db *gorm.DB) *S3BucketCORSRepositoryImpl {
	return &S3BucketCORSRepositoryImpl{db: db}
}

// GetByBucket 根据Bucket名称获取CORS配置
func (r *S3BucketCORSRepositoryImpl) GetByBucket(ctx context.Context, bucketName, userID string) (*models.S3BucketCORS, error) {
	var cors models.S3BucketCORS
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		First(&cors).Error
	if err != nil {
		return nil, err
	}
	return &cors, nil
}

// CreateOrUpdate 创建或更新CORS配置
func (r *S3BucketCORSRepositoryImpl) CreateOrUpdate(ctx context.Context, cors *models.S3BucketCORS) error {
	// 先尝试查找是否存在
	existing, err := r.GetByBucket(ctx, cors.BucketName, cors.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		cors.ID = existing.ID
		cors.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(cors).Error
	}

	// 创建
	return r.db.WithContext(ctx).Create(cors).Error
}

// Delete 删除CORS配置
func (r *S3BucketCORSRepositoryImpl) Delete(ctx context.Context, bucketName, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		Delete(&models.S3BucketCORS{}).Error
}
