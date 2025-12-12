package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type groupPowerRepository struct {
	db *gorm.DB
}

// NewGroupPowerRepository 创建组权限关联仓储实例
func NewGroupPowerRepository(db *gorm.DB) repository.GroupPowerRepository {
	return &groupPowerRepository{db: db}
}

func (r *groupPowerRepository) Create(ctx context.Context, groupPower *models.GroupPower) error {
	return r.db.WithContext(ctx).Create(groupPower).Error
}

func (r *groupPowerRepository) GetByGroupID(ctx context.Context, groupID int) ([]*models.GroupPower, error) {
	var groupPowers []*models.GroupPower
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&groupPowers).Error
	return groupPowers, err
}

func (r *groupPowerRepository) Delete(ctx context.Context, groupID, powerID int) error {
	return r.db.WithContext(ctx).Where("group_id = ? AND power_id = ?", groupID, powerID).
		Delete(&models.GroupPower{}).Error
}

func (r *groupPowerRepository) DeleteByGroupID(ctx context.Context, groupID int) error {
	return r.db.WithContext(ctx).Where("group_id = ?", groupID).Delete(&models.GroupPower{}).Error
}

func (r *groupPowerRepository) BatchCreate(ctx context.Context, groupPowers []*models.GroupPower) error {
	return r.db.WithContext(ctx).CreateInBatches(groupPowers, 100).Error
}
