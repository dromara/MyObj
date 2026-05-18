package audit

import (
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Record 异步写入审计日志
func Record(db *gorm.DB, log *models.AuditLog) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = custom_type.Now()
	}
	go func() {
		db.Create(log)
	}()
}
