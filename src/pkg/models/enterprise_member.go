package models

import (
	"myobj/src/pkg/custom_type"
)

// EnterpriseMember 企业成员
type EnterpriseMember struct {
	ID           string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	EnterpriseID string               `gorm:"type:VARCHAR;not null;index:idx_em_enterprise_id" json:"enterprise_id"`
	UserID       string               `gorm:"type:VARCHAR;not null;index:idx_em_user_id" json:"user_id"`
	RoleID       string               `gorm:"type:VARCHAR;not null" json:"role_id"`
	JoinedAt     custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"joined_at"`
	Status       int                  `gorm:"type:INTEGER;not null;default:0" json:"status"`
}

func (EnterpriseMember) TableName() string {
	return "enterprise_member"
}
