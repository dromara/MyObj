package request

// CreateSharedDirRequest 创建共享目录请求
type CreateSharedDirRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	ParentID     int    `json:"parent_id"`
}

// SharedFileListRequest 共享空间文件列表请求
type SharedFileListRequest struct {
	EnterpriseID string `form:"enterprise_id" binding:"required"`
	PathID       int    `form:"path_id"`
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=100"`
}

// DeleteSharedFileRequest 删除共享文件请求
type DeleteSharedFileRequest struct {
	ID string `json:"id" binding:"required"`
}

// SharedUploadPrecheckRequest 共享空间上传预检请求
type SharedUploadPrecheckRequest struct {
	EnterpriseID   string `json:"enterprise_id" binding:"required"`
	FileName       string `json:"file_name" binding:"required"`
	FileSize       int64  `json:"file_size" binding:"required"`
	ChunkSignature string `json:"chunk_signature"`
	PathID         int    `json:"path_id"`
}

// SharedFileUploadRequest 共享空间文件上传请求
type SharedFileUploadRequest struct {
	EnterpriseID string `form:"enterprise_id" binding:"required"`
	PathID       int    `form:"path_id"`
	PrecheckID   string `form:"precheck_id" binding:"required"`
	ChunkIndex   *int   `form:"chunk_index"`
	TotalChunks  *int   `form:"total_chunks"`
	ChunkMD5     string `form:"chunk_md5"`
	IsEnc        bool   `form:"is_enc"`
	FilePassword string `form:"file_password"`
}

// DeleteSharedDirRequest 删除共享目录请求
type DeleteSharedDirRequest struct {
	ID int `json:"id" binding:"required"`
}

// RenameSharedFileRequest 重命名共享文件请求
type RenameSharedFileRequest struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// RenameSharedDirRequest 重命名共享目录请求
type RenameSharedDirRequest struct {
	ID   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}
