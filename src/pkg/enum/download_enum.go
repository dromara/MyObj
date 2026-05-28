package enum

type DownloadTaskState int

const (
	// DownloadTaskStateInit 初始化
	DownloadTaskStateInit DownloadTaskState = iota
	// DownloadTaskStateDownloading 下载中
	DownloadTaskStateDownloading
	// DownloadTaskStatePaused 暂停
	DownloadTaskStatePaused
	// DownloadTaskStateFinished 完成
	DownloadTaskStateFinished
	// DownloadTaskStateFailed 失败
	DownloadTaskStateFailed
)

func (s DownloadTaskState) Value() int {
	return int(s)
}

type DownloadTaskType int

const (
	// DownloadTaskTypeHttp http 下载
	DownloadTaskTypeHttp DownloadTaskType = iota
	// DownloadTaskTypeFTP ftp 下载
	DownloadTaskTypeFTP
	// DownloadTaskTypeSFTP sftp 下载
	DownloadTaskTypeSFTP
	// DownloadTaskTypeS3 s3 下载
	DownloadTaskTypeS3
	// DownloadTaskTypeBtp 种子下载
	DownloadTaskTypeBtp
	// DownloadTaskTypeMagnet 磁力下载
	DownloadTaskTypeMagnet
	// DownloadTaskTypeLocal 本地下载
	DownloadTaskTypeLocal
	// DownloadTaskTypeLocalFile 网盘文件下载
	DownloadTaskTypeLocalFile
	// DownloadTaskTypePackage 打包下载
	DownloadTaskTypePackage
	// DownloadTaskTypeQuark 夸克网盘下载
	DownloadTaskTypeQuark
	// DownloadTaskTypeBaidu 百度网盘下载
	DownloadTaskTypeBaidu
	// DownloadTaskTypeAliyun 阿里云盘下载
	DownloadTaskTypeAliyun
	// DownloadTaskTypeUC UC网盘下载
	DownloadTaskTypeUC
	// DownloadTaskTypeXunlei 迅雷网盘下载
	DownloadTaskTypeXunlei
	// DownloadTaskType139 移动云盘下载
	DownloadTaskType139
	// DownloadTaskTypeTianyi 天翼云盘下载
	DownloadTaskTypeTianyi
	// DownloadTaskTypeLanzou 蓝奏云分享链接下载
	DownloadTaskTypeLanzou
	// DownloadTaskTypePan115 115网盘下载
	DownloadTaskTypePan115
	// DownloadTaskTypeOneDrive Microsoft OneDrive
	DownloadTaskTypeOneDrive
	// DownloadTaskTypeGoogleDrive Google Drive
	DownloadTaskTypeGoogleDrive
	// DownloadTaskTypeDropbox Dropbox
	DownloadTaskTypeDropbox
	// DownloadTaskTypePan123 123云盘下载
	DownloadTaskTypePan123
	// DownloadTaskTypePikPak PikPak 下载
	DownloadTaskTypePikPak
	// DownloadTaskTypeBaiduShare 百度网盘分享下载
	DownloadTaskTypeBaiduShare
	// DownloadTaskTypeAliyunShare 阿里云盘分享下载
	DownloadTaskTypeAliyunShare
	// DownloadTaskTypePan115Share 115网盘分享下载
	DownloadTaskTypePan115Share
)

func (t DownloadTaskType) Value() int {
	return int(t)
}
