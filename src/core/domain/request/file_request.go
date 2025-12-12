package request

// UploadPrecheckRequest 上传预检查
type UploadPrecheckRequest struct {
	UserID string `json:"user_id"`
	// 文件名
	FileName string `json:"file_name"`
	// 文件大小 字节
	FileSize int64 `json:"file_size"`
	// 文件hash签名
	ChunkSignature string `json:"chunk_signature"`
	// 文件分片hash 第一
	FirstChunkHash string `json:"first_chunk_hash"`
	// 文件分片hash 第二
	SecondChunkHash string `json:"second_chunk_hash"`
	// 文件分片hash 第三
	ThirdChunkHash string `json:"third_chunk_hash"`
}

// FileSearchRequest 文件搜索请求
type FileSearchRequest struct {
	Keyword  string `form:"keyword" binding:"required"`
	Type     string `form:"type"`
	SortBy   string `form:"sortBy"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	// 虚拟路径（当前所在目录）
	VirtualPath string `form:"virtualPath"`
	// 文件类型
	Type string `form:"type"`
	// 排序字段（name, size, time）
	SortBy string `form:"sortBy"`
	// 页码（从1开始）
	Page int `form:"page" binding:"required,min=1"`
	// 每页数量
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// MakeDirRequest 创建文件夹请求
type MakeDirRequest struct {
	// 父级路径
	ParentLevel string `json:"parent_level"`
	// 新文件夹路径
	DirPath string `json:"dir_path"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	// 源文件ID
	FileID string `json:"file_id"`
	// 源文件路径
	SourcePath string `json:"source_path"`
	// 目标文件路径
	TargetPath string `json:"target_path"`
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	// 文件ID列表
	FileIDs []string `json:"file_ids" binding:"required"`
}
