package cloudsync

import (
	"myobj/src/config"
	"myobj/src/pkg/cloudsync/oauth"
)

// OAuth 类 Provider 默认凭据（可由 config.toml [cloud] 覆盖）
var (
	AliyunClientID     string
	AliyunClientSecret string
	BaiduClientID      = "hq9yQ9w9kR4YHj1kyYafLygVocobh7Sf"
	BaiduClientSecret  = "YH2VpZcFJHYNnV6vLfHQXDBhcE7ZChyE"
	XunleiClientID     = "ZUBzD9J_XPXfn7f7"
	XunleiClientSecret = "yESVmHecEe6F0aou69vl-g"
	Pan115ClientID     = "100195125"
	Pan115ClientSecret string
	Pan123ClientID     string
	Pan123ClientSecret string
	PikPakClientID     string
	PikPakClientSecret string
)

// InitFromConfig 从全局配置加载云盘 OAuth 凭据
func InitFromConfig(cfg *config.MyObjConfig) {
	if cfg == nil {
		return
	}
	c := cfg.Cloud
	if c.AliyunClientID != "" {
		AliyunClientID = c.AliyunClientID
	}
	if c.AliyunClientSecret != "" {
		AliyunClientSecret = c.AliyunClientSecret
	}
	if c.BaiduClientID != "" {
		BaiduClientID = c.BaiduClientID
	}
	if c.BaiduClientSecret != "" {
		BaiduClientSecret = c.BaiduClientSecret
	}
	if c.XunleiClientID != "" {
		XunleiClientID = c.XunleiClientID
	}
	if c.XunleiClientSecret != "" {
		XunleiClientSecret = c.XunleiClientSecret
	}
	if c.Pan115ClientID != "" {
		Pan115ClientID = c.Pan115ClientID
	}
	if c.Pan115ClientSecret != "" {
		Pan115ClientSecret = c.Pan115ClientSecret
	}
	if c.Pan123ClientID != "" {
		Pan123ClientID = c.Pan123ClientID
	}
	if c.Pan123ClientSecret != "" {
		Pan123ClientSecret = c.Pan123ClientSecret
	}
	if c.PikPakClientID != "" {
		PikPakClientID = c.PikPakClientID
	}
	if c.PikPakClientSecret != "" {
		PikPakClientSecret = c.PikPakClientSecret
	}
	oauth.LoadFromConfig(c.OAuth)
}
