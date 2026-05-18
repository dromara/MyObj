package models

import (
	"myobj/src/pkg/custom_type"
)

// AuditLog 审计日志
type AuditLog struct {
	ID           string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	UserID       string               `gorm:"type:VARCHAR;not null;index:idx_audit_user_id" json:"user_id"`
	UserName     string               `gorm:"type:VARCHAR;not null" json:"user_name"`
	EnterpriseID string               `gorm:"type:VARCHAR;index:idx_audit_enterprise_id" json:"enterprise_id"`
	Action       string               `gorm:"type:VARCHAR;not null;index:idx_audit_action" json:"action"`
	TargetType   string               `gorm:"type:VARCHAR;not null" json:"target_type"`
	TargetPath   string               `gorm:"type:VARCHAR" json:"target_path"`
	TargetName   string               `gorm:"type:VARCHAR" json:"target_name"`
	Detail       string               `gorm:"type:TEXT" json:"detail"`
	IP           string               `gorm:"type:VARCHAR" json:"ip"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null;index:idx_audit_created_at" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_log"
}
