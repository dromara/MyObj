package request

// CreateOfflineDownloadRequest 创建离线下载任务请求
type CreateOfflineDownloadRequest struct {
	// 下载URL
	URL string `json:"url" binding:"required"`
	// 保存的虚拟路径（可选，默认为/离线下载/）
	VirtualPath string `json:"virtual_path"`
	// 是否加密存储
	EnableEncryption bool `json:"enable_encryption"`
	// 文件密码（加密文件必需）
	FilePassword string `json:"file_password"`
}

// DownloadTaskListRequest 下载任务列表请求
type DownloadTaskListRequest struct {
	// 任务状态（可选，0=初始化,1=下载中,2=暂停,3=完成,4=失败，-1=所有状态）
	State int `form:"state"`
	// 任务类型（可选，0-6=离线下载，7=网盘文件下载，-1=所有类型）
	// 注意：如果同时指定了 Type 和 TypeMax，优先使用 Type（单个类型查询）
	Type int `form:"type"`
	// 任务类型最大值（可选，用于范围查询，查询 type < TypeMax 的任务）
	// 例如：TypeMax=7 表示查询 type < 7 的任务（即 type 0-6，离线下载任务）
	TypeMax int `form:"typeMax"`
	// 页码
	Page int `form:"page" binding:"required,min=1"`
	// 每页数量
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// TaskOperationRequest 任务操作请求（暂停、恢复、取消）
type TaskOperationRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`
}

// DeleteTaskRequest 删除任务请求
type DeleteTaskRequest struct {
	// 任务ID
	TaskID string `json:"task_id" binding:"required"`
}

// CreateLocalFileDownloadRequest 创建网盘文件下载任务请求
type CreateLocalFileDownloadRequest struct {
	// 文件ID
	FileID string `json:"file_id" binding:"required"`
	// 文件解密密码（加密文件必需）
	FilePassword string `json:"file_password"`
}

// CreateVideoPlayRequest 创建视频播放任务请求
type CreateVideoPlayRequest struct {
	// 视频文件ID
	FileID string `json:"file_id" binding:"required"`
	// 视频文件解密密码（加密文件必需）
	FilePassword string `json:"file_password"`
}

// ParseTorrentRequest 解析种子/磁力链请求
type ParseTorrentRequest struct {
	// 种子文件内容（Base64编码）或磁力链接（magnet:开头）
	Content string `json:"content" binding:"required"`
}

// StartTorrentDownloadRequest 开始种子/磁力链下载请求
type StartTorrentDownloadRequest struct {
	// 种子文件内容（Base64编码）或磁力链接
	Content string `json:"content" binding:"required"`
	// 要下载的文件索引列表
	FileIndexes []int `json:"file_indexes" binding:"required"`
	// 保存的虚拟路径（可选，默认为/离线下载/）
	VirtualPath string `json:"virtual_path"`
	// 是否加密存储
	EnableEncryption bool `json:"enable_encryption"`
	// 文件密码（加密文件必需）
	FilePassword string `json:"file_password"`
}
