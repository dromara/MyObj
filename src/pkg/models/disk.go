package models

// Disk 磁盘信息
type Disk struct {
	ID       string `gorm:"type:varchar(64);not null;primaryKey;unique" json:"id"`          // 磁盘ID，主键且唯一
	Size     int    `gorm:"type:integer;not null" json:"size"`                              // 磁盘总大小
	DiskPath string `gorm:"type:text;not null;index:disk_disk_path_index" json:"disk_path"` // 磁盘路径
	DataPath string `gorm:"type:text;not null" json:"data_path"`                            // 数据存储路径
}

func (Disk) TableName() string {
	return "disk"
}
