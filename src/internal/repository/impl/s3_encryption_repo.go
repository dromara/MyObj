package impl

import (
	"context"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"gorm.io/gorm"
)

// S3EncryptionKeyRepositoryImpl S3加密密钥仓储实现
type S3EncryptionKeyRepositoryImpl struct {
	db *gorm.DB
}

// NewS3EncryptionKeyRepository 创建S3加密密钥仓储实例
func NewS3EncryptionKeyRepository(db *gorm.DB) *S3EncryptionKeyRepositoryImpl {
	return &S3EncryptionKeyRepositoryImpl{db: db}
}

// GetByKeyID 根据密钥ID获取加密密钥
func (r *S3EncryptionKeyRepositoryImpl) GetByKeyID(ctx context.Context, keyID string) (*models.S3EncryptionKey, error) {
	var key models.S3EncryptionKey
	err := r.db.WithContext(ctx).
		Where("key_id = ?", keyID).
		First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// Create 创建加密密钥
func (r *S3EncryptionKeyRepositoryImpl) Create(ctx context.Context, key *models.S3EncryptionKey) error {
	if key.CreatedAt.IsZero() {
		key.CreatedAt = custom_type.Now()
	}
	if key.UpdatedAt.IsZero() {
		key.UpdatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(key).Error
}

// S3ObjectEncryptionRepositoryImpl S3对象加密元数据仓储实现
type S3ObjectEncryptionRepositoryImpl struct {
	db *gorm.DB
}

// NewS3ObjectEncryptionRepository 创建S3对象加密元数据仓储实例
func NewS3ObjectEncryptionRepository(db *gorm.DB) *S3ObjectEncryptionRepositoryImpl {
	return &S3ObjectEncryptionRepositoryImpl{db: db}
}

// GetByObject 根据对象获取加密元数据
func (r *S3ObjectEncryptionRepositoryImpl) GetByObject(ctx context.Context, bucketName, objectKey, versionID, userID string) (*models.S3ObjectEncryption, error) {
	var encryption models.S3ObjectEncryption
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID)

	if versionID != "" {
		query = query.Where("version_id = ?", versionID)
	} else {
		query = query.Where("version_id = '' OR version_id IS NULL")
	}

	err := query.First(&encryption).Error
	if err != nil {
		return nil, err
	}
	return &encryption, nil
}

// Create 创建对象加密元数据
func (r *S3ObjectEncryptionRepositoryImpl) Create(ctx context.Context, encryption *models.S3ObjectEncryption) error {
	if encryption.CreatedAt.IsZero() {
		encryption.CreatedAt = custom_type.Now()
	}
	if encryption.UpdatedAt.IsZero() {
		encryption.UpdatedAt = custom_type.Now()
	}
	return r.db.WithContext(ctx).Create(encryption).Error
}

// Delete 删除对象加密元数据
func (r *S3ObjectEncryptionRepositoryImpl) Delete(ctx context.Context, bucketName, objectKey, versionID, userID string) error {
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID)

	if versionID != "" {
		query = query.Where("version_id = ?", versionID)
	} else {
		query = query.Where("version_id = '' OR version_id IS NULL")
	}

	return query.Delete(&models.S3ObjectEncryption{}).Error
}
