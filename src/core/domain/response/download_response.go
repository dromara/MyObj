package response

import "myobj/src/pkg/custom_type"

// DownloadTaskResponse 下载任务响应
type DownloadTaskResponse struct {
	// 任务ID
	ID string `json:"id"`
	// 下载URL
	URL string `json:"url"`
	// 文件名
	FileName string `json:"file_name"`
	// 文件大小
	FileSize int64 `json:"file_size"`
	// 已下载大小
	DownloadedSize int64 `json:"downloaded_size"`
	// 下载进度（0-100）
	Progress int `json:"progress"`
	// 下载速度（字节/秒）
	Speed int64 `json:"speed"`
	// 任务类型（0=HTTP, 1=FTP, 2=SFTP, 3=S3, 4=BT, 5=磁力, 6=本地）
	Type int `json:"type"`
	// 类型文本
	TypeText string `json:"type_text"`
	// 任务状态（0=初始化,1=下载中,2=暂停,3=完成,4=失败）
	State int `json:"state"`
	// 状态文本
	StateText string `json:"state_text"`
	// 虚拟路径
	VirtualPath string `json:"virtual_path"`
	// 是否支持断点续传
	SupportRange bool `json:"support_range"`
	// 错误信息
	ErrorMsg string `json:"error_msg"`
	// 文件ID（下载完成后）
	FileID string `json:"file_id"`
	// 创建时间
	CreateTime custom_type.JsonTime `json:"create_time"`
	// 更新时间
	UpdateTime custom_type.JsonTime `json:"update_time"`
	// 完成时间
	FinishTime custom_type.JsonTime `json:"finish_time"`
}

// DownloadTaskListResponse 下载任务列表响应
type DownloadTaskListResponse struct {
	// 任务列表
	Tasks []*DownloadTaskResponse `json:"tasks"`
	// 总数
	Total int64 `json:"total"`
	// 当前页
	Page int `json:"page"`
	// 每页数量
	PageSize int `json:"page_size"`
}

// TorrentFileInfo 种子文件信息
type TorrentFileInfo struct {
	// 文件索引
	Index int `json:"index"`
	// 文件名
	Name string `json:"name"`
	// 文件大小（字节）
	Size int64 `json:"size"`
	// 文件路径（种子内的相对路径）
	Path string `json:"path"`
}

// ParseTorrentResponse 解析种子/磁力链响应
type ParseTorrentResponse struct {
	// 种子名称
	Name string `json:"name"`
	// InfoHash（种子唯一标识）
	InfoHash string `json:"info_hash"`
	// 文件列表
	Files []TorrentFileInfo `json:"files"`
	// 总大小（字节）
	TotalSize int64 `json:"total_size"`
}

// StartTorrentDownloadResponse 开始种子下载响应
type StartTorrentDownloadResponse struct {
	// 创建的任务ID列表
	TaskIDs []string `json:"task_ids"`
	// 种子名称
	TorrentName string `json:"torrent_name"`
	// 创建的任务数量
	TaskCount int `json:"task_count"`
}
