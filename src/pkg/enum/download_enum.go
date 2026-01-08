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
)

func (t DownloadTaskType) Value() int {
	return int(t)
}
