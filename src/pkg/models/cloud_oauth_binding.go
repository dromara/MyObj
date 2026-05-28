package models

import "myobj/src/pkg/custom_type"

// CloudOAuthBinding 云盘 OAuth 绑定（国际网盘授权）
type CloudOAuthBinding struct {
	ID           string               `gorm:"type:VARCHAR(36);primaryKey" json:"id"`
	UserID       string               `gorm:"type:VARCHAR(36);not null;index" json:"user_id"`
	Provider     string               `gorm:"type:VARCHAR(32);not null;index" json:"provider"`
	AccessToken  string               `gorm:"type:TEXT;not null" json:"-"`
	RefreshToken string               `gorm:"type:TEXT" json:"-"`
	ExpiresAt    custom_type.JsonTime `gorm:"type:DATETIME" json:"expires_at"`
	AccountName  string               `gorm:"type:VARCHAR(128)" json:"account_name"`
	ExtraData    string               `gorm:"type:TEXT" json:"-"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (CloudOAuthBinding) TableName() string {
	return "cloud_oauth_binding"
}
