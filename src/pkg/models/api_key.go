package models

import (
	"myobj/src/pkg/custom_type"
)

// ApiKey API密钥
type ApiKey struct {
	ID          int                  `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"` // API密钥ID，主键且唯一
	UserID      string               `gorm:"type:VARCHAR;not null" json:"user_id"`              // 用户ID
	Key         string               `gorm:"type:TEXT;not null" json:"key"`                     // API密钥
	ExpiresAt   custom_type.JsonTime `gorm:"type:DATETIME" json:"expires_at"`                   // 过期时间
	CreatedAt   custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`          // 创建时间
	PrivateKey  string               `gorm:"type:text;not null" json:"private_key"`             // 私钥（RSA，用于加密/解密）
	S3SecretKey string               `gorm:"type:text;not null" json:"s3_secret_key"`            // S3 Secret Key（字符串，用于HMAC-SHA256签名）
}

func (ApiKey) TableName() string {
	return "api_key"
}
