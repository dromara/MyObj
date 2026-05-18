package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseMemberRepository struct {
	db *gorm.DB
}

func NewEnterpriseMemberRepository(db *gorm.DB) repository.EnterpriseMemberRepository {
	return &enterpriseMemberRepository{db: db}
}

func (r *enterpriseMemberRepository) Create(ctx context.Context, member *models.EnterpriseMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *enterpriseMemberRepository) GetByID(ctx context.Context, id string) (*models.EnterpriseMember, error) {
	var member models.EnterpriseMember
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&member).Error
	return &member, err
}

func (r *enterpriseMemberRepository) GetByEnterpriseAndUser(ctx context.Context, enterpriseID, userID string) (*models.EnterpriseMember, error) {
	var member models.EnterpriseMember
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND user_id = ? AND status = 0", enterpriseID, userID).First(&member).Error
	return &member, err
}

func (r *enterpriseMemberRepository) ListByEnterpriseID(ctx context.Context, enterpriseID string, offset, limit int) ([]*models.EnterpriseMember, error) {
	var members []*models.EnterpriseMember
	err := r.db.WithContext(ctx).Where("enterprise_id = ? AND status = 0", enterpriseID).Offset(offset).Limit(limit).Find(&members).Error
	return members, err
}

func (r *enterpriseMemberRepository) CountByEnterpriseID(ctx context.Context, enterpriseID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseMember{}).Where("enterprise_id = ? AND status = 0", enterpriseID).Count(&count).Error
	return count, err
}

func (r *enterpriseMemberRepository) ListByUserID(ctx context.Context, userID string) ([]*models.EnterpriseMember, error) {
	var members []*models.EnterpriseMember
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = 0", userID).Find(&members).Error
	return members, err
}

func (r *enterpriseMemberRepository) Update(ctx context.Context, member *models.EnterpriseMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *enterpriseMemberRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.EnterpriseMember{}).Error
}
