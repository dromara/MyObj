package response

// SharesDownloadResponse 分享下载响应
type SharesDownloadResponse struct {
	Path     string `json:"path"`
	Temp     string `json:"temp"`
	Err      string `json:"err"`
	FileName string `json:"file_name"`
}

// ShareInfoResponse 分享信息响应（不触发下载）
type ShareInfoResponse struct {
	FileID        string `json:"file_id"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	HasPassword   bool   `json:"has_password"`   // 是否有密码
	ExpiresAt     string `json:"expires_at"`     // 过期时间
	DownloadCount int    `json:"download_count"` // 下载次数
	IsExpired     bool   `json:"is_expired"`     // 是否已过期
}
