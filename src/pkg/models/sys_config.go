package models

// SysConfig 系统配置表
type SysConfig struct {
	ID    int    `gorm:"type:INTEGER;not null;primaryKey;unique" json:"id"`
	Key   string `gorm:"type:VARCHAR;not null" json:"key"`
	Value string `gorm:"type:TEXT;not null" json:"value"`
}

func (SysConfig) TableName() string {
	return "sys_config"
}
