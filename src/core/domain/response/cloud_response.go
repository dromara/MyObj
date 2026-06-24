package response

import "time"

// ShareLinkInfoResponse 分享链接信息响应
type ShareLinkInfoResponse struct {
	TaskID     int              `json:"task_id"`
	ShareID    string           `json:"share_id"`
	ShareTitle string           `json:"share_title"`
	Provider   string           `json:"provider"`
	FileCount  int              `json:"file_count"`
	TotalSize  int64            `json:"total_size"`
	Files      []ShareFileInfo  `json:"files"`
	ExpiresAt  *time.Time       `json:"expires_at"`
	Status     int              `json:"status"`
}

// ShareFileInfo 分享文件信息
type ShareFileInfo struct {
	FileID    string    `json:"file_id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	IsDir     bool      `json:"is_dir"`
	FileType  string    `json:"file_type"`
	FileExt   string    `json:"file_ext"`
	UpdatedAt time.Time `json:"updated_at"`
	Thumbnail string    `json:"thumbnail,omitempty"`
}

// ShareFileListResponse 分享文件列表响应
type ShareFileListResponse struct {
	TaskID       int              `json:"task_id"`
	ParentFileID string           `json:"parent_file_id"`
	Files        []ShareFileInfo  `json:"files"`
	HasMore      bool             `json:"has_more"`
}

// ShareTaskStatusResponse 分享任务状态响应
type ShareTaskStatusResponse struct {
	TaskID         int        `json:"task_id"`
	Provider       string     `json:"provider"`
	ShareID        string     `json:"share_id"`
	ShareTitle     string     `json:"share_title"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	FileCount      int        `json:"file_count"`
	TotalSize      int64      `json:"total_size"`
	SuccessCount   int        `json:"success_count"`
	FailedCount    int        `json:"failed_count"`
	DownloadCount  int        `json:"download_count"`
	TargetPath     string     `json:"target_path"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Progress       float64    `json:"progress"` // 下载进度 0-100
}

// ShareTaskListResponse 分享任务列表响应
type ShareTaskListResponse struct {
	Total int                    `json:"total"`
	Tasks []ShareTaskStatusResponse `json:"tasks"`
}
