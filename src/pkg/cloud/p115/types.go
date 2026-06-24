package p115

// ShareSnapRequest 获取分享快照请求
type ShareSnapRequest struct {
	ShareCode string `json:"share_code"`
	ReceiveID string `json:"receive_id,omitempty"`
}

// ShareSnapResponse 获取分享快照响应
type ShareSnapResponse struct {
	ShareID      string          `json:"share_id"`
	ShareName    string          `json:"share_name"`
	ShareTitle   string          `json:"share_title"`
	ShareSize    int64           `json:"share_size"`
	FileCount    int             `json:"file_count"`
	FolderCount  int             `json:"folder_count"`
	CreateTime   int64           `json:"create_time"`
	ExpiredTime  int64           `json:"expired_time"`
	IsExpired    bool            `json:"is_expired"`
	HasPassword  bool            `json:"has_password"`
	ReceiveCode  string          `json:"receive_code"`
	Items        []ShareSnapItem `json:"items"`
}

// ShareSnapItem 分享快照中的文件/目录项
type ShareSnapItem struct {
	FileID       string `json:"file_id"`
	ParentID     string `json:"parent_id"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	IsDir        bool   `json:"is_dir"`
	PickCode     string `json:"pick_code"`
	Sha1         string `json:"sha1"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
	ThumbURL     string `json:"thumb_url"`
	FileCategory int    `json:"file_category"`
}

// ShareSnapDirRequest 获取分享目录下文件列表请求
type ShareSnapDirRequest struct {
	ShareCode string `json:"share_code"`
	DirID     string `json:"dir_id"`
	ReceiveID string `json:"receive_id,omitempty"`
}

// ShareSnapDirResponse 获取分享目录下文件列表响应
type ShareSnapDirResponse struct {
	Items []ShareSnapItem `json:"items"`
}

// DownloadURLResponse 获取下载链接响应
type DownloadURLResponse struct {
	URL string `json:"url"`
}

// ShareDownloadRequest 获取分享下载链接请求
type ShareDownloadRequest struct {
	ShareCode string `json:"share_code"`
	FileID    string `json:"file_id"`
	PickCode  string `json:"pick_code"`
}

// ShareDownloadResponse 获取分享下载链接响应
type ShareDownloadResponse struct {
	FileID   string                   `json:"file_id"`
	FileName string                   `json:"file_name"`
	FileSize int64                    `json:"file_size"`
	Urls     []ShareDownloadURLItem   `json:"urls"`
}

// ShareDownloadURLItem 下载链接项
type ShareDownloadURLItem struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsVip    bool   `json:"is_vip"`
}

// APIResponse 通用API响应
type APIResponse struct {
	State   bool   `json:"state"`
	ErrNo   int    `json:"errno"`
	Error   string `json:"error"`
	Message string `json:"message"`
}
