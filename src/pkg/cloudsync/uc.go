package cloudsync

func init() {
	RegisterProvider("uc", func(cookie string) CloudProvider {
		return NewQuarkProviderWithConfig(cookie, QuarkConfig{
			API:     "https://pc-api.uc.cn/1/clouddrive",
			Referer: "https://drive.uc.cn",
			UA:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) uc-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
			PR:      "UCBrowser",
		})
	})
}
