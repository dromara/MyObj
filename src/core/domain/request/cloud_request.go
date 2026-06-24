package request

// ParseShareLinkRequest 解析分享链接请求
type ParseShareLinkRequest struct {
	// 云盘提供者: aliyun, baidu, xunlei, 115, quark (可选，支持自动检测)
	Provider string `json:"provider"`
	// 分享链接
	ShareURL string `json:"share_url" binding:"required"`
	// 提取码（可选）
	SharePwd string `json:"share_pwd"`
	// 目标存储路径（可选，默认为根目录）
	TargetPath string `json:"target_path"`
}

// ListShareFilesRequest 列出分享文件请求
type ListShareFilesRequest struct {
	// 任务ID
	TaskID int `json:"task_id" binding:"required"`
	// 父目录ID（为空则列出根目录）
	ParentFileID string `json:"parent_file_id"`
}

// DownloadShareFileRequest 下载分享文件请求
type DownloadShareFileRequest struct {
	// 任务ID
	TaskID int `json:"task_id" binding:"required"`
	// 文件ID列表（为空则下载所有文件）
	FileIDs []string `json:"file_ids"`
	// 目标路径（可选）
	TargetPath string `json:"target_path"`
}

// GetShareTaskStatusRequest 获取任务状态请求
type GetShareTaskStatusRequest struct {
	// 任务ID
	TaskID int `json:"task_id" binding:"required"`
}
