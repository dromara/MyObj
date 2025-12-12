package models

// UploadChunk 上传分片信息
type UploadChunk struct {
	ChunkID  int    `json:"chunk_id" gorm:"primaryKey"`
	UserID   string `json:"user_id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	Md5      string `json:"md5"`
	PathID   string `json:"path_id"`
}

func (UploadChunk) TableName() string {
	return "upload_chunk"
}
