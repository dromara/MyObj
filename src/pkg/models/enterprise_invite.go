package models

import (
	"myobj/src/pkg/custom_type"
)

// EnterpriseInvite 企业邀请记录
type EnterpriseInvite struct {
	ID           string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	EnterpriseID string               `gorm:"type:VARCHAR;not null;index:idx_ei_enterprise_id" json:"enterprise_id"`
	InviterID    string               `gorm:"type:VARCHAR;not null" json:"inviter_id"`
	InviteeID    string               `gorm:"type:VARCHAR" json:"invitee_id"`
	InviteCode   string               `gorm:"type:VARCHAR" json:"invite_code"`
	Type         int                  `gorm:"type:INTEGER;not null" json:"type"`
	Status       int                  `gorm:"type:INTEGER;not null;default:0" json:"status"`
	ExpireAt     custom_type.JsonTime `gorm:"type:DATETIME" json:"expire_at"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
}

func (EnterpriseInvite) TableName() string {
	return "enterprise_invite"
}
