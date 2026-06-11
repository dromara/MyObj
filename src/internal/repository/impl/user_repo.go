package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.UserInfo) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUserName(ctx context.Context, userName string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := r.db.WithContext(ctx).Where("user_name = ?", userName).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.UserInfo) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.UserInfo{}).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.UserInfo, error) {
	var users []*models.UserInfo
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserInfo{}).Count(&count).Error
	return count, err
}

// CountByGroupID 统计指定组下的用户数量
func (r *userRepository) CountByGroupID(ctx context.Context, groupID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserInfo{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

// BatchGetByIDs 根据多个ID批量查询用户信息
func (r *userRepository) BatchGetByIDs(ctx context.Context, ids []string) (map[string]*models.UserInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []*models.UserInfo
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*models.UserInfo, len(users))
	for _, u := range users {
		result[u.ID] = u
	}
	return result, nil
}
