package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type enterpriseInviteRepository struct {
	db *gorm.DB
}

func NewEnterpriseInviteRepository(db *gorm.DB) repository.EnterpriseInviteRepository {
	return &enterpriseInviteRepository{db: db}
}

func (r *enterpriseInviteRepository) Create(ctx context.Context, invite *models.EnterpriseInvite) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *enterpriseInviteRepository) GetByID(ctx context.Context, id string) (*models.EnterpriseInvite, error) {
	var invite models.EnterpriseInvite
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&invite).Error
	return &invite, err
}

func (r *enterpriseInviteRepository) ListPendingByInviteeID(ctx context.Context, inviteeID string) ([]*models.EnterpriseInvite, error) {
	var invites []*models.EnterpriseInvite
	err := r.db.WithContext(ctx).Where("invitee_id = ? AND status = 0", inviteeID).Find(&invites).Error
	return invites, err
}

func (r *enterpriseInviteRepository) ListByEnterpriseID(ctx context.Context, enterpriseID string, offset, limit int) ([]*models.EnterpriseInvite, error) {
	var invites []*models.EnterpriseInvite
	query := r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Order("created_at DESC")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	err := query.Find(&invites).Error
	return invites, err
}

func (r *enterpriseInviteRepository) CountByEnterpriseID(ctx context.Context, enterpriseID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EnterpriseInvite{}).Where("enterprise_id = ?", enterpriseID).Count(&count).Error
	return count, err
}

func (r *enterpriseInviteRepository) Update(ctx context.Context, invite *models.EnterpriseInvite) error {
	return r.db.WithContext(ctx).Save(invite).Error
}
