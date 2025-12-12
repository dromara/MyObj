package response

// SharesDownloadResponse 分享下载响应
type SharesDownloadResponse struct {
	Path     string `json:"path"`
	Temp     string `json:"temp"`
	Err      string `json:"err"`
	FileName string `json:"file_name"`
}
