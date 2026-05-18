package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseRolePowerRepository struct {
	db *gorm.DB
}

func NewEnterpriseRolePowerRepository(db *gorm.DB) repository.EnterpriseRolePowerRepository {
	return &enterpriseRolePowerRepository{db: db}
}

func (r *enterpriseRolePowerRepository) BatchCreate(ctx context.Context, rolePowers []*models.EnterpriseRolePower) error {
	return r.db.WithContext(ctx).Create(rolePowers).Error
}

func (r *enterpriseRolePowerRepository) DeleteByRoleID(ctx context.Context, roleID string) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.EnterpriseRolePower{}).Error
}

func (r *enterpriseRolePowerRepository) GetByRoleID(ctx context.Context, roleID string) ([]*models.EnterpriseRolePower, error) {
	var rolePowers []*models.EnterpriseRolePower
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rolePowers).Error
	return rolePowers, err
}
