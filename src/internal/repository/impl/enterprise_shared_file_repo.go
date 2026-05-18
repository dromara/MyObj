package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseSharedFileRepository struct {
	db *gorm.DB
}

func NewEnterpriseSharedFileRepository(db *gorm.DB) repository.EnterpriseSharedFileRepository {
	return &enterpriseSharedFileRepository{db: db}
}

func (r *enterpriseSharedFileRepository) Create(ctx context.Context, file *models.EnterpriseSharedFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *enterpriseSharedFileRepository) GetByID(ctx context.Context, id string) (*models.EnterpriseSharedFile, error) {
	var file models.EnterpriseSharedFile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error
	return &file, err
}

func (r *enterpriseSharedFileRepository) ListByPathID(ctx context.Context, enterpriseID string, pathID int, offset, limit int) ([]*models.EnterpriseSharedFile, error) {
	var files []*models.EnterpriseSharedFile
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND path_id = ?", enterpriseID, pathID).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *enterpriseSharedFileRepository) CountByPathID(ctx context.Context, enterpriseID string, pathID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).
		Where("enterprise_id = ? AND path_id = ?", enterpriseID, pathID).Count(&count).Error
	return count, err
}

func (r *enterpriseSharedFileRepository) CountByEnterpriseID(ctx context.Context, enterpriseID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).
		Where("enterprise_id = ?", enterpriseID).Count(&count).Error
	return count, err
}

func (r *enterpriseSharedFileRepository) SumSizeByEnterpriseID(ctx context.Context, enterpriseID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).
		Where("enterprise_id = ?", enterpriseID).Select("COALESCE(SUM(size), 0)").Scan(&total).Error
	return total, err
}

func (r *enterpriseSharedFileRepository) Update(ctx context.Context, file *models.EnterpriseSharedFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *enterpriseSharedFileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.EnterpriseSharedFile{}).Error
}

func (r *enterpriseSharedFileRepository) DeleteByEnterpriseID(ctx context.Context, enterpriseID string) error {
	return r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Delete(&models.EnterpriseSharedFile{}).Error
}
