package cloudsync

// CloudFile 云盘文件信息
type CloudFile struct {
	Fid      string `json:"fid"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"is_dir"`
}

// CloudDownloadLink 云盘下载链接
type CloudDownloadLink struct {
	DownloadURL string            `json:"download_url"`
	FileName    string            `json:"file_name"`
	Size        int64             `json:"size"`
	Headers     map[string]string `json:"-"` // 下载时需要携带的请求头（如 Cookie、Referer）
}

// CloudUserInfo 云盘用户信息
type CloudUserInfo struct {
	Nickname  string `json:"nickname"`
	TotalSize int64  `json:"total_size"`
	UsedSize  int64  `json:"used_size"`
}
