package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseRepository struct {
	db *gorm.DB
}

func NewEnterpriseRepository(db *gorm.DB) repository.EnterpriseRepository {
	return &enterpriseRepository{db: db}
}

func (r *enterpriseRepository) Create(ctx context.Context, enterprise *models.Enterprise) error {
	return r.db.WithContext(ctx).Create(enterprise).Error
}

func (r *enterpriseRepository) GetByID(ctx context.Context, id string) (*models.Enterprise, error) {
	var enterprise models.Enterprise
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&enterprise).Error
	return &enterprise, err
}

func (r *enterpriseRepository) GetByInviteCode(ctx context.Context, code string) (*models.Enterprise, error) {
	var enterprise models.Enterprise
	err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&enterprise).Error
	return &enterprise, err
}

func (r *enterpriseRepository) Update(ctx context.Context, enterprise *models.Enterprise) error {
	return r.db.WithContext(ctx).Save(enterprise).Error
}

func (r *enterpriseRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Enterprise{}).Error
}
