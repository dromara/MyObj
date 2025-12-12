package models

// FileChunk 文件分片信息
type FileChunk struct {
	// 分片ID
	ID string `gorm:"type:VARCHAR;not null;primaryKey;unique" json:"id"`
	// 文件ID
	FileID string `gorm:"type:VARCHAR;not null" json:"file_id"`
	// 分片文件路径
	ChunkPath string `gorm:"type:TEXT;not null" json:"chunk_path"`
	// 分片文件大小
	ChunkSize uint64 `gorm:"type:BIGINT;not null" json:"chunk_size"`
	// 分片文件哈希
	ChunkHash string `gorm:"type:TEXT;not null" json:"chunk_hash"`
	// 分片文件索引
	ChunkIndex uint32 `gorm:"type:INTEGER;not null" json:"chunk_index"`
}

func (FileChunk) TableName() string {
	return "file_chunk"
}
