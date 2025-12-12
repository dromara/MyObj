package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type fileChunkRepository struct {
	db *gorm.DB
}

// NewFileChunkRepository 创建文件分片仓储实例
func NewFileChunkRepository(db *gorm.DB) repository.FileChunkRepository {
	return &fileChunkRepository{db: db}
}

func (r *fileChunkRepository) Create(ctx context.Context, chunk *models.FileChunk) error {
	return r.db.WithContext(ctx).Create(chunk).Error
}

func (r *fileChunkRepository) GetByID(ctx context.Context, id string) (*models.FileChunk, error) {
	var chunk models.FileChunk
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (r *fileChunkRepository) GetByFileID(ctx context.Context, fileID string) ([]*models.FileChunk, error) {
	var chunks []*models.FileChunk
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

func (r *fileChunkRepository) Update(ctx context.Context, chunk *models.FileChunk) error {
	return r.db.WithContext(ctx).Save(chunk).Error
}

func (r *fileChunkRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.FileChunk{}).Error
}

func (r *fileChunkRepository) DeleteByFileID(ctx context.Context, fileID string) error {
	return r.db.WithContext(ctx).Where("file_id = ?", fileID).Delete(&models.FileChunk{}).Error
}

func (r *fileChunkRepository) BatchCreate(ctx context.Context, chunks []*models.FileChunk) error {
	return r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error
}
