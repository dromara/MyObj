package models

import (
	"myobj/src/pkg/custom_type"
)

type FileInfo struct {
	ID              string               `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`                      // 文件ID，主键且唯一
	Name            string               `gorm:"type:VARCHAR;not null;index:file_info_index_1" json:"name"`              // 文件原名
	RandomName      string               `gorm:"type:VARCHAR;not null" json:"random_name"`                               // 文件存储名（随机生成）
	Size            int                  `gorm:"type:INTEGER;not null" json:"size"`                                      // 文件大小
	Mime            string               `gorm:"type:VARCHAR;not null;index:file_info_index_0" json:"mime"`              // 文件MIME类型
	ThumbnailImg    string               `gorm:"type:TEXT" json:"thumbnail_img"`                                         // 缩略图路径
	Path            string               `gorm:"type:TEXT" json:"path"`                                                  // 文件实际存储路径
	FileHash        string               `gorm:"type:TEXT;not null;index:file_info_file_hash_index" json:"file_hash"`    // 文件哈希值（全量hash）
	FileEncHash     string               `gorm:"type:TEXT" json:"file_enc_hash"`                                         // 加密文件哈希值
	ChunkSignature  string               `gorm:"type:TEXT;index:file_info_chunk_signature_index" json:"chunk_signature"` // 分片签名（快速预检）
	FirstChunkHash  string               `gorm:"type:TEXT" json:"first_chunk_hash"`                                      // 第一个分片hash
	SecondChunkHash string               `gorm:"type:TEXT" json:"second_chunk_hash"`                                     // 第二个分片hash
	ThirdChunkHash  string               `gorm:"type:TEXT" json:"third_chunk_hash"`                                      // 第三个分片hash
	HasFullHash     bool                 `gorm:"type:BOOLEAN;default:false" json:"has_full_hash"`                        // 是否已计算全量hash
	IsEnc           bool                 `gorm:"type:BOOLEAN" json:"is_enc"`                                             // 是否加密
	IsChunk         bool                 `gorm:"type:BOOLEAN;not null" json:"is_chunk"`                                  // 是否分块存储
	ChunkCount      int                  `gorm:"type:INTEGER" json:"chunk_count"`                                        // 分块数量
	EncPath         string               `gorm:"type:TEXT;not null" json:"enc_path"`                                     // 加密文件路径
	CreatedAt       custom_type.JsonTime `gorm:"type:DATETIME" json:"created_at"`                                        // 创建时间
	UpdatedAt       custom_type.JsonTime `gorm:"type:DATETIME" json:"updated_at"`                                        // 更新时间
}

func (FileInfo) TableName() string {
	return "file_info"
}
