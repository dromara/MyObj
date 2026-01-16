package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type powerRepository struct {
	db *gorm.DB
}

// NewPowerRepository 创建权限仓储实例
func NewPowerRepository(db *gorm.DB) repository.PowerRepository {
	return &powerRepository{db: db}
}

func (r *powerRepository) Create(ctx context.Context, power *models.Power) error {
	return r.db.WithContext(ctx).Create(power).Error
}

func (r *powerRepository) GetByID(ctx context.Context, id int) (*models.Power, error) {
	var power models.Power
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&power).Error
	if err != nil {
		return nil, err
	}
	return &power, nil
}

func (r *powerRepository) Update(ctx context.Context, power *models.Power) error {
	return r.db.WithContext(ctx).Save(power).Error
}

func (r *powerRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Power{}).Error
}

func (r *powerRepository) List(ctx context.Context, offset, limit int) ([]*models.Power, error) {
	var powers []*models.Power
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&powers).Error
	return powers, err
}

func (r *powerRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Power{}).Count(&count).Error
	return count, err
}

func (r *powerRepository) GetByGroupID(ctx context.Context, groupID int) ([]*models.Power, error) {
	var powers []*models.Power
	err := r.db.WithContext(ctx).Model(&models.Power{}).
		Joins("LEFT JOIN group_power ON power.id = group_power.power_id").
		Joins("LEFT JOIN `groups` ON group_power.group_id = `groups`.id").
		Where("`groups`.id = ?", groupID).
		Find(&powers).Error
	return powers, err
}
