package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseRoleRepository struct {
	db *gorm.DB
}

func NewEnterpriseRoleRepository(db *gorm.DB) repository.EnterpriseRoleRepository {
	return &enterpriseRoleRepository{db: db}
}

func (r *enterpriseRoleRepository) Create(ctx context.Context, role *models.EnterpriseRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *enterpriseRoleRepository) GetByID(ctx context.Context, id string) (*models.EnterpriseRole, error) {
	var role models.EnterpriseRole
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	return &role, err
}

func (r *enterpriseRoleRepository) GetDefaultByEnterpriseID(ctx context.Context, enterpriseID string) (*models.EnterpriseRole, error) {
	var role models.EnterpriseRole
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND is_default = 1", enterpriseID).First(&role).Error
	return &role, err
}

func (r *enterpriseRoleRepository) GetAdminByEnterpriseID(ctx context.Context, enterpriseID string) (*models.EnterpriseRole, error) {
	var role models.EnterpriseRole
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND is_admin = 1", enterpriseID).First(&role).Error
	return &role, err
}

func (r *enterpriseRoleRepository) ListByEnterpriseID(ctx context.Context, enterpriseID string) ([]*models.EnterpriseRole, error) {
	var roles []*models.EnterpriseRole
	err := r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Find(&roles).Error
	return roles, err
}

func (r *enterpriseRoleRepository) Update(ctx context.Context, role *models.EnterpriseRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *enterpriseRoleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.EnterpriseRole{}).Error
}
