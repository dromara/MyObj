package impl

import (
	"context"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"gorm.io/gorm"
)

// S3ACLRepositoryImpl S3 ACL配置仓储实现
type S3ACLRepositoryImpl struct {
	db *gorm.DB
}

// NewS3ACLRepository 创建S3 ACL配置仓储实例
func NewS3ACLRepository(db *gorm.DB) *S3ACLRepositoryImpl {
	return &S3ACLRepositoryImpl{db: db}
}

// ==================== Bucket ACL ====================

// GetBucketACL 根据Bucket名称获取ACL配置
func (r *S3ACLRepositoryImpl) GetBucketACL(ctx context.Context, bucketName, userID string) (*models.S3BucketACL, error) {
	var acl models.S3BucketACL
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		First(&acl).Error
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

// CreateOrUpdateBucketACL 创建或更新Bucket ACL配置
func (r *S3ACLRepositoryImpl) CreateOrUpdateBucketACL(ctx context.Context, acl *models.S3BucketACL) error {
	// 先尝试查找是否存在
	existing, err := r.GetBucketACL(ctx, acl.BucketName, acl.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		acl.ID = existing.ID
		acl.CreatedAt = existing.CreatedAt
		if acl.CreatedAt.IsZero() {
			acl.CreatedAt = existing.CreatedAt
		}
		return r.db.WithContext(ctx).Save(acl).Error
	}

	// 创建
	if acl.CreatedAt.IsZero() {
		acl.CreatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(acl).Error
}

// DeleteBucketACL 删除Bucket ACL配置
func (r *S3ACLRepositoryImpl) DeleteBucketACL(ctx context.Context, bucketName, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID).
		Delete(&models.S3BucketACL{}).Error
}

// ==================== Object ACL ====================

// GetObjectACL 根据Object Key获取ACL配置
func (r *S3ACLRepositoryImpl) GetObjectACL(ctx context.Context, bucketName, objectKey, versionID, userID string) (*models.S3ObjectACL, error) {
	var acl models.S3ObjectACL
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID)

	if versionID != "" {
		query = query.Where("version_id = ?", versionID)
	} else {
		query = query.Where("(version_id = '' OR version_id IS NULL)")
	}

	err := query.First(&acl).Error
	if err != nil {
		return nil, err
	}
	return &acl, nil
}

// CreateOrUpdateObjectACL 创建或更新Object ACL配置
func (r *S3ACLRepositoryImpl) CreateOrUpdateObjectACL(ctx context.Context, acl *models.S3ObjectACL) error {
	// 先尝试查找是否存在
	existing, err := r.GetObjectACL(ctx, acl.BucketName, acl.ObjectKey, acl.VersionID, acl.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existing != nil {
		// 更新
		acl.ID = existing.ID
		acl.CreatedAt = existing.CreatedAt
		if acl.CreatedAt.IsZero() {
			acl.CreatedAt = existing.CreatedAt
		}
		return r.db.WithContext(ctx).Save(acl).Error
	}

	// 创建
	if acl.CreatedAt.IsZero() {
		acl.CreatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(acl).Error
}

// DeleteObjectACL 删除Object ACL配置
func (r *S3ACLRepositoryImpl) DeleteObjectACL(ctx context.Context, bucketName, objectKey, versionID, userID string) error {
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID)

	if versionID != "" {
		query = query.Where("version_id = ?", versionID)
	} else {
		query = query.Where("(version_id = '' OR version_id IS NULL)")
	}

	return query.Delete(&models.S3ObjectACL{}).Error
}
