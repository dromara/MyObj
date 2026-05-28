package provider

import (
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/enum"
)

func init() {
	cloudsync.RegisterShareLinkProvider(cloudsync.ProviderInfo{
		ID:          "lanzou",
		Name:        "蓝奏云",
		AuthType:    cloudsync.AuthShareLink,
		Description: "通过分享链接解析下载，免费用户单文件通常不超过 100MB",
		MaxFileSize: 100 * 1024 * 1024,
	}, enum.DownloadTaskTypeLanzou.Value())
}
