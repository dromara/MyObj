package impl

import (
	"context"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) repository.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) ListByCondition(ctx context.Context, query *repository.AuditLogQuery) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&models.AuditLog{})

	if query.EnterpriseID != "" {
		db = db.Where("enterprise_id = ?", query.EnterpriseID)
	}
	if query.UserID != "" {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Keyword != "" {
		db = db.Where("target_name LIKE ? OR target_path LIKE ? OR detail LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
