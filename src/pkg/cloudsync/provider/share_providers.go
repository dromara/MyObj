package provider

import (
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/enum"
	_ "myobj/src/pkg/sharelink"
)

func init() {
	cloudsync.RegisterShareLinkProvider(cloudsync.ProviderInfo{
		ID:          "baidu_share",
		Name:        "百度网盘分享",
		AuthType:    cloudsync.AuthShareLink,
		Description: "解析百度网盘分享链接，需要 BDUSS Cookie",
	}, enum.DownloadTaskTypeBaiduShare.Value())

	cloudsync.RegisterShareLinkProvider(cloudsync.ProviderInfo{
		ID:          "aliyun_share",
		Name:        "阿里云盘分享",
		AuthType:    cloudsync.AuthShareLink,
		Description: "解析阿里云盘分享链接，需要用户 refresh_token",
	}, enum.DownloadTaskTypeAliyunShare.Value())

	cloudsync.RegisterShareLinkProvider(cloudsync.ProviderInfo{
		ID:          "115_share",
		Name:        "115网盘分享",
		AuthType:    cloudsync.AuthShareLink,
		Description: "解析115网盘分享链接，需要115账号 Cookie",
	}, enum.DownloadTaskTypePan115Share.Value())
}
