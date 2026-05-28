package cloudsync

import "time"

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
	Headers     map[string]string `json:"headers,omitempty"`
	MustProxy   bool              `json:"must_proxy,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// CloudUserInfo 云盘用户信息
type CloudUserInfo struct {
	Nickname  string `json:"nickname"`
	TotalSize int64  `json:"total_size"`
	UsedSize  int64  `json:"used_size"`
}

// CredentialField 凭据表单字段（供前端自动生成输入项）
type CredentialField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret,omitempty"`
	Help     string `json:"help,omitempty"`
}
