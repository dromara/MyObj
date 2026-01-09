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
	// 路径ID
	PathID string `json:"path_id"`
	// 文件分片的DM5列表
	FilesMd5 []string `json:"files_md5"`
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

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	// 预检ID
	PrecheckID string `form:"precheck_id" binding:"required"`
	// 分片索引（分片上传时必须，从0开始）
	ChunkIndex *int `form:"chunk_index"`
	// 总分片数（分片上传时必须）
	TotalChunks *int `form:"total_chunks"`
	// 当前分片的MD5（分片上传时必须）
	ChunkMD5 string `form:"chunk_md5"`
	// 是否需要加密
	IsEnc bool `form:"is_enc"`
	// 文件加密密码（加密文件必须）
	FilePassword string `form:"file_password"`
}

// VideoPlayPrecheckRequest 视频播放预检请求
type VideoPlayPrecheckRequest struct {
	// 文件ID
	FileID string `json:"file_id" binding:"required"`
	// 分享密码（如果是分享链接访问）
	SharePassword string `json:"share_password"`
}

// PublicFileListRequest 公开文件列表请求
type PublicFileListRequest struct {
	// 文件类型
	Type string `form:"type"`
	// 排序字段（name, size, time）
	SortBy string `form:"sortBy"`
	// 页码（从1开始）
	Page int `form:"page" binding:"required,min=1"`
	// 每页数量
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// UploadProgressRequest 上传进度查询请求
type UploadProgressRequest struct {
	// 预检ID
	PrecheckID string `form:"precheck_id" binding:"required"`
}

// DeleteUploadTaskRequest 删除上传任务请求
type DeleteUploadTaskRequest struct {
	// 任务ID（预检ID）
	TaskID string `json:"task_id" binding:"required"`
}

// RenewExpiredTaskRequest 延期过期任务请求
type RenewExpiredTaskRequest struct {
	// 任务ID（预检ID）
	TaskID string `json:"task_id" binding:"required"`
	// 延期天数（默认7天）
	Days int `json:"days"`
}

// RenameFileRequest 文件重命名请求
type RenameFileRequest struct {
	// 文件ID（uf_id）
	FileID string `json:"file_id" binding:"required"`
	// 新文件名
	NewFileName string `json:"new_file_name" binding:"required"`
}

// RenameDirRequest 目录重命名请求
type RenameDirRequest struct {
	// 目录ID
	DirID int `json:"dir_id" binding:"required"`
	// 新目录名
	NewDirName string `json:"new_dir_name" binding:"required"`
}

// SetFilePublicRequest 设置文件公开状态请求
type SetFilePublicRequest struct {
	// 文件ID（uf_id）
	FileID string `json:"file_id" binding:"required"`
	// 是否公开
	Public bool `json:"public"`
}

// DeleteDirRequest 删除目录请求
type DeleteDirRequest struct {
	// 目录ID
	DirID int `json:"dir_id" binding:"required"`
}

// UploadTaskListRequest 上传任务列表请求
type UploadTaskListRequest struct {
	// 页码（从1开始）
	Page int `form:"page" binding:"required,min=1"`
	// 每页数量
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}
