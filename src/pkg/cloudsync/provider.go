package cloudsync

import "fmt"

// CloudProvider 云盘提供者接口
// 接入新网盘只需实现此接口并调用 RegisterProvider 注册即可
type CloudProvider interface {
	// Name 返回提供者名称（如 "quark", "baidu", "aliyun"）
	Name() string

	// ListFiles 列出目录下的文件
	ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error)

	// GetDownloadLink 获取文件的临时下载链接
	GetDownloadLink(fid string) (*CloudDownloadLink, error)

	// Validate 验证凭据是否有效，返回用户信息
	Validate() (*CloudUserInfo, error)
}

var providers = make(map[string]func(cookie string) CloudProvider)

// RegisterProvider 注册云盘提供者工厂函数
func RegisterProvider(name string, factory func(cookie string) CloudProvider) {
	providers[name] = factory
}

// GetProvider 根据名称获取云盘提供者
func GetProvider(name, cookie string) (CloudProvider, error) {
	factory, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", name)
	}
	return factory(cookie), nil
}
