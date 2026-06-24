package baidu

// ShareVerifyRequest 验证分享链接请求
type ShareVerifyRequest struct {
	Surl string `json:"surl"`
	Pwd  string `json:"pwd"`
}

// ShareVerifyResponse 验证分享链接响应
type ShareVerifyResponse struct {
	Errno     int    `json:"errno"`
	ErrMsg    string `json:"errmsg"`
	ShareID   string `json:"shareid"`
	Uk        string `json:"uk"`
	Bdstoken  string `json:"bdstoken"`
}

// ShareFileListRequest 获取分享文件列表请求
type ShareFileListRequest struct {
	ShareID string `json:"shareid"`
	Uk      string `json:"uk"`
	Page    int    `json:"page"`
	Num     int    `json:"num"`
	Order   string `json:"order"`
}

// ShareFileListResponse 获取分享文件列表响应
type ShareFileListResponse struct {
	Errno     int              `json:"errno"`
	ErrMsg    string           `json:"errmsg"`
	List      []ShareFileItem  `json:"list"`
	Total     int              `json:"total"`
	Page      int              `json:"page"`
	Num       int              `json:"num"`
}

// ShareFileItem 分享文件项
type ShareFileItem struct {
	FsID        int64  `json:"fs_id"`
	Path        string `json:"path"`
	Filename    string `json:"filename"`
	Isdir       int    `json:"isdir"` // 0=文件, 1=目录
	Size        int64  `json:"size"`
	ServerMtime int64  `json:"server_mtime"` // 修改时间戳
	ServerCtime int64  `json:"server_ctime"` // 创建时间戳
	Md5         string `json:"md5"`
	Category    int    `json:"category"` // 文件类别
}

// ShareDownloadRequest 获取下载链接请求
type ShareDownloadRequest struct {
	ShareID string `json:"shareid"`
	Uk      string `json:"uk"`
	FsID    int64  `json:"fs_id"`
}

// ShareDownloadResponse 获取下载链接响应
type ShareDownloadResponse struct {
	Errno  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
	Dlink  string `json:"dlink"`
	Info   []struct {
		FsID     int64  `json:"fs_id"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		MD5      string `json:"md5"`
		Dlink    string `json:"dlink"`
	} `json:"info"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Errno  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
}