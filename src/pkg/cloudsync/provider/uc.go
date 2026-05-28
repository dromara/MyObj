package provider

import (
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/enum"
)

func init() {
	cloudsync.Register(cloudsync.ProviderInfo{
		ID: "uc", Name: "UC网盘", AuthType: cloudsync.AuthCookie,
		Description: "Cookie 登录，基于夸克 API",
	}, enum.DownloadTaskTypeUC.Value(), func(cookie string) cloudsync.CloudProvider {
		return NewQuarkProviderWithConfig(cookie, QuarkConfig{
			ProviderID: "uc",
			API:        "https://pc-api.uc.cn/1/clouddrive",
			Referer:    "https://drive.uc.cn",
			UA:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) uc-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
			PR:         "UCBrowser",
		})
	})
}
