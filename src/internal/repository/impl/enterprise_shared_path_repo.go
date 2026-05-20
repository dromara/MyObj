package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseSharedPathRepository struct {
	db *gorm.DB
}

func NewEnterpriseSharedPathRepository(db *gorm.DB) repository.EnterpriseSharedPathRepository {
	return &enterpriseSharedPathRepository{db: db}
}

func (r *enterpriseSharedPathRepository) Create(ctx context.Context, path *models.EnterpriseSharedPath) error {
	return r.db.WithContext(ctx).Create(path).Error
}

func (r *enterpriseSharedPathRepository) GetByID(ctx context.Context, id int) (*models.EnterpriseSharedPath, error) {
	var path models.EnterpriseSharedPath
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&path).Error
	if err != nil {
		return nil, err
	}
	return &path, nil
}

func (r *enterpriseSharedPathRepository) GetByParentIDAndName(ctx context.Context, enterpriseID string, parentID int, name string) (*models.EnterpriseSharedPath, error) {
	var path models.EnterpriseSharedPath
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND parent_id = ? AND name = ?", enterpriseID, parentID, name).First(&path).Error
	if err != nil {
		return nil, err
	}
	return &path, nil
}

func (r *enterpriseSharedPathRepository) ListByParentID(ctx context.Context, enterpriseID string, parentID int) ([]*models.EnterpriseSharedPath, error) {
	var paths []*models.EnterpriseSharedPath
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND parent_id = ?", enterpriseID, parentID).Find(&paths).Error
	return paths, err
}

func (r *enterpriseSharedPathRepository) GetPathTree(ctx context.Context, enterpriseID string) ([]*models.EnterpriseSharedPath, error) {
	var paths []*models.EnterpriseSharedPath
	err := r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Order("parent_id, name").Find(&paths).Error
	return paths, err
}

func (r *enterpriseSharedPathRepository) Update(ctx context.Context, path *models.EnterpriseSharedPath) error {
	return r.db.WithContext(ctx).Model(&models.EnterpriseSharedPath{}).Where("id = ?", path.ID).
		Select("name", "parent_id", "created_by", "updated_by", "created_at", "updated_at").
		Updates(path).Error
}

func (r *enterpriseSharedPathRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.EnterpriseSharedPath{}).Error
}

func (r *enterpriseSharedPathRepository) DeleteByEnterpriseID(ctx context.Context, enterpriseID string) error {
	return r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Delete(&models.EnterpriseSharedPath{}).Error
}
