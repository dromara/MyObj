package models

import "myobj/src/pkg/custom_type"

// CloudCredentialBinding 云盘凭据绑定（Cookie / refresh_token 等）
type CloudCredentialBinding struct {
	ID          string               `gorm:"type:VARCHAR(36);primaryKey" json:"id"`
	UserID      string               `gorm:"type:VARCHAR(36);not null;index" json:"user_id"`
	Provider    string               `gorm:"type:VARCHAR(32);not null;index" json:"provider"`
	Credential  string               `gorm:"type:TEXT;not null" json:"-"`
	AccountName string               `gorm:"type:VARCHAR(128)" json:"account_name"`
	CreatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"updated_at"`
}

func (CloudCredentialBinding) TableName() string {
	return "cloud_credential_binding"
}
