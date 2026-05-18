package models

import (
	"myobj/src/pkg/custom_type"
)

// Enterprise 企业信息
type Enterprise struct {
	ID          string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	Name        string               `gorm:"type:VARCHAR;not null" json:"name"`
	Logo        string               `gorm:"type:TEXT" json:"logo"`
	Description string               `gorm:"type:TEXT" json:"description"`
	CreatorID   string               `gorm:"type:VARCHAR;not null" json:"creator_id"`
	Space       int64                `gorm:"type:BIGINT;not null;default:0" json:"space"`
	FreeSpace   int64                `gorm:"type:BIGINT;not null;default:0" json:"free_space"`
	InviteCode  string               `gorm:"type:VARCHAR;unique" json:"invite_code"`
	InviteLink  string               `gorm:"type:TEXT" json:"invite_link"`
	State       int                  `gorm:"type:INTEGER;not null;default:0" json:"state"`
	CreatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
}

func (Enterprise) TableName() string {
	return "enterprise"
}
