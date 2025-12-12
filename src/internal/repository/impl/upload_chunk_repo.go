package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type uploadChunkRepository struct {
	db *gorm.DB
}

// NewUploadChunkRepository 创建上传分片信息仓储实例
func NewUploadChunkRepository(db *gorm.DB) repository.UploadChunkRepository {
	return &uploadChunkRepository{db: db}
}

// Create 创建上传分片记录
func (r *uploadChunkRepository) Create(ctx context.Context, chunk *models.UploadChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

// GetByID 根据分片ID获取上传分片信息
func (r *uploadChunkRepository) GetByID(ctx context.Context, chunkID int) (*models.UploadChunk, error) {
	var chunk models.UploadChunk
	err := r.db.WithContext(ctx).Where("chunk_id = ?", chunkID).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

// Update 更新上传分片信息
func (r *uploadChunkRepository) Update(ctx context.Context, chunk *models.UploadChunk) error {
	return r.db.WithContext(ctx).Save(chunk).Error
}

// Delete 删除上传分片记录
func (r *uploadChunkRepository) Delete(ctx context.Context, chunkID int) error {
	return r.db.WithContext(ctx).Where("chunk_id = ?", chunkID).Delete(&models.UploadChunk{}).Error
}

// ListByUserID 根据用户ID获取上传分片列表
func (r *uploadChunkRepository) ListByUserID(ctx context.Context, userID string, offset, limit int) ([]*models.UploadChunk, error) {
	var chunks []*models.UploadChunk
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Offset(offset).Limit(limit).Find(&chunks).Error
	return chunks, err
}

// Count 统计用户的上传分片数量
func (r *uploadChunkRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UploadChunk{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// GetByUserIDAndFileName 根据用户ID和文件名获取分片信息
func (r *uploadChunkRepository) GetByUserIDAndFileName(ctx context.Context, userID, fileName string) ([]models.UploadChunk, error) {
	var chunk []models.UploadChunk
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND file_name = ?", userID, fileName).
		Find(&chunk).Error
	if err != nil {
		return nil, err
	}
	return chunk, nil
}

// DeleteByUserID 删除用户的所有上传分片记录
func (r *uploadChunkRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UploadChunk{}).Error
}

// ListByPathID 根据路径ID获取分片列表
func (r *uploadChunkRepository) ListByPathID(ctx context.Context, pathID string, offset, limit int) ([]*models.UploadChunk, error) {
	var chunks []*models.UploadChunk
	err := r.db.WithContext(ctx).Where("path_id = ?", pathID).
		Offset(offset).Limit(limit).Find(&chunks).Error
	return chunks, err
}
