package cloudsync

// CloudProvider 云盘 Provider 接口（可插拔）
//
// 包结构：
//   - cloudsync/         注册表、类型、配置入口
//   - cloudsync/internal  HTTP/凭据/分页等工具
//   - cloudsync/provider 各网盘实现（init 注册）
//   - cloudsync/oauth    国际网盘 OAuth 配置
//
// 接入新网盘步骤：
//  1. 在 provider/ 下实现本接口
//  2. 在 init() 中调用 Register 注册元数据、任务类型与工厂函数
//  3. 在 download_enum.go 增加对应 DownloadTaskType（如尚未存在）
//
// 离线下载场景使用 pdirFid 作为目录标识即可，无需挂载路径或 VFS 层。
type CloudProvider interface {
	Name() string
	ListFiles(pdirFid string, page, size int) ([]CloudFile, int, error)
	GetDownloadLink(fid string) (*CloudDownloadLink, error)
	Validate() (*CloudUserInfo, error)
}
