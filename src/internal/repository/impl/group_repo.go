package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

// NewGroupRepository 创建组仓储实例
func NewGroupRepository(db *gorm.DB) repository.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) GetByID(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) Update(ctx context.Context, group *models.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *groupRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Group{}).Error
}

func (r *groupRepository) List(ctx context.Context, offset, limit int) ([]*models.Group, error) {
	var groups []*models.Group
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&groups).Error
	return groups, err
}

func (r *groupRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Group{}).Count(&count).Error
	return count, err
}

func (r *groupRepository) GetDefaultGroup(ctx context.Context) (*models.Group, error) {
	var group models.Group
	err := r.db.WithContext(ctx).Where("group_default = ?", 1).First(&group).Error
	return &group, err
}
