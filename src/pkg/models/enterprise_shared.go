package models

import (
	"myobj/src/pkg/custom_type"
)

// EnterpriseSharedPath 企业共享空间目录
type EnterpriseSharedPath struct {
	ID           int                  `gorm:"type:INTEGER;not null;primaryKey;unique;autoIncrement" json:"id"`
	EnterpriseID string               `gorm:"type:VARCHAR;not null;index:idx_esp_enterprise_id" json:"enterprise_id"`
	Name         string               `gorm:"type:VARCHAR;not null" json:"name"`
	ParentID     int                  `gorm:"type:INTEGER;not null;default:0" json:"parent_id"`
	CreatedBy    string               `gorm:"type:VARCHAR" json:"created_by"`
	UpdatedBy    string               `gorm:"type:VARCHAR" json:"updated_by"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt    custom_type.JsonTime `gorm:"type:DATETIME" json:"updated_at"`
}

func (EnterpriseSharedPath) TableName() string {
	return "enterprise_shared_path"
}

// EnterpriseSharedFile 企业共享空间文件关联
type EnterpriseSharedFile struct {
	ID           string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	EnterpriseID string               `gorm:"type:VARCHAR;not null;index:idx_esf_enterprise_id" json:"enterprise_id"`
	FileID       string               `gorm:"type:VARCHAR;not null;index:idx_esf_file_id" json:"file_id"`
	FileName     string               `gorm:"type:VARCHAR;not null" json:"file_name"`
	PathID       int                  `gorm:"type:INTEGER;not null;default:0" json:"path_id"`
	UploaderID   string               `gorm:"type:VARCHAR;not null" json:"uploader_id"`
	Size         int64                `gorm:"type:BIGINT;not null" json:"size"`
	UpdatedBy    string               `gorm:"type:VARCHAR" json:"updated_by"`
	CreatedAt    custom_type.JsonTime `gorm:"type:DATETIME;not null" json:"created_at"`
	UpdatedAt    custom_type.JsonTime `gorm:"type:DATETIME" json:"updated_at"`
}

func (EnterpriseSharedFile) TableName() string {
	return "enterprise_shared_file"
}
