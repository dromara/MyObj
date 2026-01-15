package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"time"

	"gorm.io/gorm"
)

// S3ObjectMetadataRepositoryImpl S3对象元数据仓储实现
type S3ObjectMetadataRepositoryImpl struct {
	db *gorm.DB
}

// NewS3ObjectMetadataRepository 创建S3对象元数据仓储实例
func NewS3ObjectMetadataRepository(db *gorm.DB) repository.S3ObjectMetadataRepository {
	return &S3ObjectMetadataRepositoryImpl{db: db}
}

// Create 创建对象元数据
func (r *S3ObjectMetadataRepositoryImpl) Create(ctx context.Context, metadata *models.S3ObjectMetadata) error {
	return r.db.WithContext(ctx).Create(metadata).Error
}

// GetByKey 根据Bucket和Key获取对象元数据
func (r *S3ObjectMetadataRepositoryImpl) GetByKey(ctx context.Context, bucketName, objectKey, userID string) (*models.S3ObjectMetadata, error) {
	var metadata models.S3ObjectMetadata
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ? AND is_latest = ?", bucketName, objectKey, userID, true).
		First(&metadata).Error
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// ListByBucket 列出Bucket下的对象
func (r *S3ObjectMetadataRepositoryImpl) ListByBucket(ctx context.Context, bucketName, userID, prefix string, maxKeys int, marker string) ([]*models.S3ObjectMetadata, error) {
	var metadatas []*models.S3ObjectMetadata
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ? AND is_latest = ?", bucketName, userID, true)

	if prefix != "" {
		query = query.Where("object_key LIKE ?", prefix+"%")
	}

	if marker != "" {
		query = query.Where("object_key > ?", marker)
	}

	err := query.Order("object_key ASC").Limit(maxKeys).Find(&metadatas).Error
	return metadatas, err
}

// CountByBucket 统计Bucket下的对象数量
func (r *S3ObjectMetadataRepositoryImpl) CountByBucket(ctx context.Context, bucketName, userID, prefix string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.S3ObjectMetadata{}).
		Where("bucket_name = ? AND user_id = ? AND is_latest = ?", bucketName, userID, true)

	if prefix != "" {
		query = query.Where("object_key LIKE ?", prefix+"%")
	}

	err := query.Count(&count).Error
	return count, err
}

// Update 更新对象元数据
func (r *S3ObjectMetadataRepositoryImpl) Update(ctx context.Context, metadata *models.S3ObjectMetadata) error {
	return r.db.WithContext(ctx).Save(metadata).Error
}

// Delete 删除对象元数据
func (r *S3ObjectMetadataRepositoryImpl) Delete(ctx context.Context, bucketName, objectKey, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID).
		Delete(&models.S3ObjectMetadata{}).Error
}

// DeleteByVersion 根据版本ID删除对象元数据
func (r *S3ObjectMetadataRepositoryImpl) DeleteByVersion(ctx context.Context, bucketName, objectKey, versionID, userID string) error {
	return r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND version_id = ? AND user_id = ?", bucketName, objectKey, versionID, userID).
		Delete(&models.S3ObjectMetadata{}).Error
}

// MarkOldVersions 标记旧版本（版本控制）
func (r *S3ObjectMetadataRepositoryImpl) MarkOldVersions(ctx context.Context, bucketName, objectKey, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.S3ObjectMetadata{}).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID).
		Update("is_latest", false).Error
}

// CountByFileID 统计指定文件被多少个S3对象引用
func (r *S3ObjectMetadataRepositoryImpl) CountByFileID(ctx context.Context, fileID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.S3ObjectMetadata{}).
		Where("file_id = ?", fileID).
		Count(&count).Error
	return count, err
}

// ListVersionsByBucket 列出Bucket下的对象版本（包括所有版本，不限于最新版本）
func (r *S3ObjectMetadataRepositoryImpl) ListVersionsByBucket(ctx context.Context, bucketName, userID, prefix, keyMarker, versionIDMarker string, maxKeys int) ([]*models.S3ObjectMetadata, error) {
	var metadatas []*models.S3ObjectMetadata
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ?", bucketName, userID)

	if prefix != "" {
		query = query.Where("object_key LIKE ?", prefix+"%")
	}

	// 处理分页标记（keyMarker 和 versionIDMarker）
	if keyMarker != "" {
		if versionIDMarker != "" {
			// 同时有 keyMarker 和 versionIDMarker
			query = query.Where("(object_key > ? OR (object_key = ? AND version_id > ?))", keyMarker, keyMarker, versionIDMarker)
		} else {
			query = query.Where("object_key > ?", keyMarker)
		}
	} else if versionIDMarker != "" {
		query = query.Where("version_id > ?", versionIDMarker)
	}

	err := query.Order("object_key ASC, version_id DESC").Limit(maxKeys + 1).Find(&metadatas).Error
	return metadatas, err
}

// GetByKeyAndVersion 根据Bucket、Key和版本ID获取对象元数据（包括DeleteMarker）
func (r *S3ObjectMetadataRepositoryImpl) GetByKeyAndVersion(ctx context.Context, bucketName, objectKey, versionID, userID string) (*models.S3ObjectMetadata, error) {
	var metadata models.S3ObjectMetadata
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ? AND user_id = ?", bucketName, objectKey, userID)

	if versionID != "" {
		query = query.Where("version_id = ?", versionID)
	} else {
		// 如果版本ID为空，获取最新版本（排除DeleteMarker）
		query = query.Where("is_latest = ? AND is_delete_marker = ?", true, false)
	}

	err := query.First(&metadata).Error
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// ListObjectVersions 列出对象的所有版本（ListObjectVersions的别名，调用ListVersionsByBucket）
func (r *S3ObjectMetadataRepositoryImpl) ListObjectVersions(ctx context.Context, bucketName, userID, prefix, keyMarker, versionIDMarker string, maxKeys int) ([]*models.S3ObjectMetadata, error) {
	return r.ListVersionsByBucket(ctx, bucketName, userID, prefix, keyMarker, versionIDMarker, maxKeys)
}

// S3MultipartRepositoryImpl S3分片上传仓储实现
type S3MultipartRepositoryImpl struct {
	db *gorm.DB
}

// NewS3MultipartRepository 创建S3分片上传仓储实例
func NewS3MultipartRepository(db *gorm.DB) repository.S3MultipartRepository {
	return &S3MultipartRepositoryImpl{db: db}
}

// CreateUpload 创建分片上传会话
func (r *S3MultipartRepositoryImpl) CreateUpload(ctx context.Context, upload *models.S3MultipartUpload) error {
	return r.db.WithContext(ctx).Create(upload).Error
}

// GetUpload 获取分片上传会话
func (r *S3MultipartRepositoryImpl) GetUpload(ctx context.Context, uploadID string) (*models.S3MultipartUpload, error) {
	var upload models.S3MultipartUpload
	err := r.db.WithContext(ctx).Where("upload_id = ?", uploadID).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

// UpdateUploadStatus 更新上传状态
func (r *S3MultipartRepositoryImpl) UpdateUploadStatus(ctx context.Context, uploadID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.S3MultipartUpload{}).
		Where("upload_id = ?", uploadID).
		Update("status", status).Error
}

// DeleteUpload 删除上传会话
func (r *S3MultipartRepositoryImpl) DeleteUpload(ctx context.Context, uploadID string) error {
	return r.db.WithContext(ctx).Where("upload_id = ?", uploadID).Delete(&models.S3MultipartUpload{}).Error
}

// CreatePart 创建分片
func (r *S3MultipartRepositoryImpl) CreatePart(ctx context.Context, part *models.S3MultipartPart) error {
	return r.db.WithContext(ctx).Create(part).Error
}

// GetPart 获取分片
func (r *S3MultipartRepositoryImpl) GetPart(ctx context.Context, uploadID string, partNumber int) (*models.S3MultipartPart, error) {
	var part models.S3MultipartPart
	err := r.db.WithContext(ctx).
		Where("upload_id = ? AND part_number = ?", uploadID, partNumber).
		First(&part).Error
	if err != nil {
		return nil, err
	}
	return &part, nil
}

// ListParts 列出所有分片
func (r *S3MultipartRepositoryImpl) ListParts(ctx context.Context, uploadID string) ([]*models.S3MultipartPart, error) {
	var parts []*models.S3MultipartPart
	err := r.db.WithContext(ctx).
		Where("upload_id = ?", uploadID).
		Order("part_number ASC").
		Find(&parts).Error
	return parts, err
}

// DeletePart 删除分片
func (r *S3MultipartRepositoryImpl) DeletePart(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.S3MultipartPart{}, id).Error
}

// DeletePartsByUploadID 删除上传会话的所有分片
func (r *S3MultipartRepositoryImpl) DeletePartsByUploadID(ctx context.Context, uploadID string) error {
	return r.db.WithContext(ctx).Where("upload_id = ?", uploadID).Delete(&models.S3MultipartPart{}).Error
}

// ListUploads 列出分片上传会话
func (r *S3MultipartRepositoryImpl) ListUploads(ctx context.Context, bucketName, userID, prefix, keyMarker, uploadIDMarker string, maxUploads int) ([]*models.S3MultipartUpload, error) {
	var uploads []*models.S3MultipartUpload
	query := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ? AND status = ?", bucketName, userID, "in-progress")

	if prefix != "" {
		query = query.Where("object_key LIKE ?", prefix+"%")
	}

	// 处理分页标记
	if keyMarker != "" {
		if uploadIDMarker != "" {
			// 同时有 keyMarker 和 uploadIDMarker
			query = query.Where("(object_key > ? OR (object_key = ? AND upload_id > ?))", keyMarker, keyMarker, uploadIDMarker)
		} else {
			query = query.Where("object_key > ?", keyMarker)
		}
	} else if uploadIDMarker != "" {
		query = query.Where("upload_id > ?", uploadIDMarker)
	}

	err := query.Order("object_key ASC, upload_id ASC").Limit(maxUploads + 1).Find(&uploads).Error
	return uploads, err
}

// ListMultipartUploadsByBucket 列出指定Bucket下的所有未完成分片上传（用于Lifecycle清理）
func (r *S3MultipartRepositoryImpl) ListMultipartUploadsByBucket(ctx context.Context, bucketName, userID string, beforeTime time.Time) ([]*models.S3MultipartUpload, error) {
	var uploads []*models.S3MultipartUpload
	err := r.db.WithContext(ctx).
		Where("bucket_name = ? AND user_id = ? AND status = ? AND created_at < ?", bucketName, userID, "in-progress", beforeTime).
		Order("created_at ASC").
		Find(&uploads).Error
	return uploads, err
}