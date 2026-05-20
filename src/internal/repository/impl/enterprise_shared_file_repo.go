package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"strings"

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
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *enterpriseSharedFileRepository) GetByPathIDAndName(ctx context.Context, enterpriseID string, pathID int, fileName string) (*models.EnterpriseSharedFile, error) {
	var file models.EnterpriseSharedFile
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND path_id = ? AND file_name = ?", enterpriseID, pathID, fileName).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *enterpriseSharedFileRepository) ListByPathID(ctx context.Context, enterpriseID string, pathID int, offset, limit int) ([]*models.EnterpriseSharedFile, error) {
	var files []*models.EnterpriseSharedFile
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND path_id = ?", enterpriseID, pathID).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *enterpriseSharedFileRepository) ListByPathIDWithSort(ctx context.Context, enterpriseID string, pathID int, sortField, sortOrder string, offset, limit int) ([]*models.EnterpriseSharedFile, error) {
	var files []*models.EnterpriseSharedFile
	orderClause := sortField + " " + sortOrder
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND path_id = ?", enterpriseID, pathID).
		Offset(offset).Limit(limit).Order(orderClause).Find(&files).Error
	return files, err
}

func (r *enterpriseSharedFileRepository) ListByEnterpriseID(ctx context.Context, enterpriseID string, keyword string, offset, limit int) ([]*models.EnterpriseSharedFile, error) {
	var files []*models.EnterpriseSharedFile
	query := r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID)
	if keyword != "" {
		// 转义 LIKE 通配符防止注入
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(keyword)
		query = query.Where("file_name LIKE ?", "%"+escaped+"%")
	}
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
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

func (r *enterpriseSharedFileRepository) CountByEnterpriseIDAndKeyword(ctx context.Context, enterpriseID string, keyword string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).Where("enterprise_id = ?", enterpriseID)
	if keyword != "" {
		escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(keyword)
		query = query.Where("file_name LIKE ?", "%"+escaped+"%")
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *enterpriseSharedFileRepository) ListByFileID(ctx context.Context, fileID string) ([]*models.EnterpriseSharedFile, error) {
	var files []*models.EnterpriseSharedFile
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).Find(&files).Error
	return files, err
}

func (r *enterpriseSharedFileRepository) SumSizeByEnterpriseID(ctx context.Context, enterpriseID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).
		Where("enterprise_id = ?", enterpriseID).Select("COALESCE(SUM(size), 0)").Scan(&total).Error
	return total, err
}

func (r *enterpriseSharedFileRepository) Update(ctx context.Context, file *models.EnterpriseSharedFile) error {
	return r.db.WithContext(ctx).Model(&models.EnterpriseSharedFile{}).Where("id = ?", file.ID).
		Select("file_name", "path_id", "uploader_id", "size", "updated_by", "updated_at").
		Updates(file).Error
}

func (r *enterpriseSharedFileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.EnterpriseSharedFile{}).Error
}

func (r *enterpriseSharedFileRepository) DeleteByEnterpriseID(ctx context.Context, enterpriseID string) error {
	return r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Delete(&models.EnterpriseSharedFile{}).Error
}
