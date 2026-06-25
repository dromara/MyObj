package models

import (
	"myobj/src/pkg/custom_type"
)

// CloudAccount 云盘账号
type CloudAccount struct {
	ID           int                   `gorm:"type:INTEGER;primaryKey;autoIncrement" json:"id"`
	UserID       string                `gorm:"type:VARCHAR(255);not null;index" json:"user_id"`
	Provider     string                `gorm:"type:VARCHAR(255);not null" json:"provider"`
	AccountName  string                `gorm:"type:VARCHAR(255)" json:"account_name"`
	AccessToken  string                `gorm:"type:TEXT" json:"-"`
	RefreshToken string                `gorm:"type:TEXT" json:"-"`
	Cookie       string                `gorm:"type:TEXT" json:"-"`
	ExpiresAt    *custom_type.JsonTime `gorm:"type:DATETIME" json:"expires_at"`
	Status       int                   `gorm:"type:INTEGER;default:1" json:"status"` // 1=有效 0=过期 -1=失效
	CreatedAt    custom_type.JsonTime  `gorm:"type:DATETIME" json:"created_at"`
	UpdatedAt    custom_type.JsonTime  `gorm:"type:DATETIME" json:"updated_at"`
}

func (CloudAccount) TableName() string {
	return "cloud_accounts"
}

const (
	CloudAccountStatusValid   = 1
	CloudAccountStatusExpired = 0
	CloudAccountStatusInvalid = -1
)
