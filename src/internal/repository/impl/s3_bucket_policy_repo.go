package impl

import (
	"context"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"gorm.io/gorm"
)

// S3BucketPolicyRepositoryImpl S3 Bucket Policy配置仓储实现
type S3BucketPolicyRepositoryImpl struct {
	db *gorm.DB
}

// NewS3BucketPolicyRepository 创建S3 Bucket Policy配置仓储实例
func NewS3BucketPolicyRepository(db *gorm.DB) *S3BucketPolicyRepositoryImpl {
	return &S3BucketPolicyRepositoryImpl{db: db}
}

// GetByBucket 根据Bucket名称获取Policy配置
func (r *S3BucketPolicyRepositoryImpl) GetByBucket(ctx context.Context, bucketName, userID string) (*models.S3BucketPolicy, error) {
	var policy models.S3BucketPolicy
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// CreateOrUpdate 创建或更新Policy配置
func (r *S3BucketPolicyRepositoryImpl) CreateOrUpdate(ctx context.Context, policy *models.S3BucketPolicy) error {
	// 先尝试查找是否存在
	existing, err := r.GetByBucket(ctx, policy.BucketName, policy.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
		if policy.CreatedAt.IsZero() {
			policy.CreatedAt = existing.CreatedAt
		}
		return r.db.WithContext(ctx).Save(policy).Error
	}

	// 创建
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(policy).Error
}

// Delete 删除Policy配置
func (r *S3BucketPolicyRepositoryImpl) Delete(ctx context.Context, bucketName, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		Delete(&models.S3BucketPolicy{}).Error
}
