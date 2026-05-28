package cloudsync

import (
	"myobj/src/pkg/enum"
	"fmt"
	"sort"
	"strings"
)

// AuthType 云盘认证方式
type AuthType string

const (
	AuthCookie       AuthType = "cookie"
	AuthRefreshToken AuthType = "refresh_token"
	AuthOAuth2       AuthType = "oauth2"
	AuthShareLink    AuthType = "share_link"
)

// ProviderInfo 云盘 Provider 元数据（供 API 与前端使用）
type ProviderInfo struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	AuthType         AuthType          `json:"auth_type"`
	Description      string            `json:"description,omitempty"`
	MaxFileSize      int64             `json:"max_file_size,omitempty"`
	Enabled          bool              `json:"enabled"`
	RequiresProxy    bool              `json:"requires_proxy,omitempty"`
	CredentialFields []CredentialField `json:"credential_fields,omitempty"`
}

type providerEntry struct {
	info      ProviderInfo
	taskType  int
	factory   func(credential string) CloudProvider
}

var providerRegistry = make(map[string]*providerEntry)

// Register 注册云盘 Provider（工厂 + 元数据 + 任务类型）
func Register(info ProviderInfo, taskType int, factory func(credential string) CloudProvider) {
	info.Enabled = true
	if len(info.CredentialFields) == 0 {
		info.CredentialFields = defaultCredentialFields(info.AuthType)
	}
	providerRegistry[info.ID] = &providerEntry{
		info:     info,
		taskType: taskType,
		factory:  factory,
	}
}

func defaultCredentialFields(auth AuthType) []CredentialField {
	switch auth {
	case AuthCookie:
		return []CredentialField{
			{Key: "cookie", Label: "浏览器 Cookie", Required: true, Secret: true, Help: "从浏览器开发者工具复制完整 Cookie"},
		}
	case AuthRefreshToken:
		return []CredentialField{
			{Key: "refresh_token", Label: "刷新令牌 (refresh_token)", Required: true, Secret: true, Help: "粘贴开放平台 refresh_token"},
			{Key: "client_id", Label: "应用 ID (Client ID)", Required: false, Help: "留空则使用 config.toml [cloud] 配置"},
			{Key: "client_secret", Label: "应用密钥 (Client Secret)", Required: false, Secret: true, Help: "留空则使用 config.toml [cloud] 配置"},
		}
	case AuthOAuth2:
		return nil
	default:
		return nil
	}
}

// GetProvider 根据 ID 获取 Provider 实例
func GetProvider(name, credential string) (CloudProvider, error) {
	entry, ok := providerRegistry[name]
	if !ok || !entry.info.Enabled {
		return nil, errUnsupportedProvider(name)
	}
	if entry.factory == nil {
		return nil, errUnsupportedProvider(name)
	}
	return entry.factory(credential), nil
}

// ListProviders 返回所有已启用的 Provider 元数据
func ListProviders() []ProviderInfo {
	list := make([]ProviderInfo, 0, len(providerRegistry))
	for _, entry := range providerRegistry {
		if entry.info.Enabled {
			list = append(list, entry.info)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return list
}

// GetTaskType 根据 Provider ID 获取下载任务类型
func GetTaskType(providerID string) int {
	if entry, ok := providerRegistry[providerID]; ok {
		return entry.taskType
	}
	return enum.DownloadTaskTypeHttp.Value()
}

// GetProviderInfo 获取 Provider 元数据
func GetProviderInfo(providerID string) (ProviderInfo, bool) {
	entry, ok := providerRegistry[providerID]
	if !ok || !entry.info.Enabled {
		return ProviderInfo{}, false
	}
	return entry.info, true
}

// CheckCredential 校验 Provider 与凭据基础参数
func CheckCredential(providerID, credential string) error {
	if strings.TrimSpace(providerID) == "" {
		return fmt.Errorf("云盘类型不能为空")
	}
	if strings.TrimSpace(credential) == "" {
		return fmt.Errorf("凭据不能为空")
	}
	info, ok := GetProviderInfo(providerID)
	if !ok {
		return errUnsupportedProvider(providerID)
	}
	if info.AuthType == AuthShareLink {
		return fmt.Errorf("该云盘不支持凭据登录，请使用分享链接接口")
	}
	if info.AuthType == AuthOAuth2 {
		return fmt.Errorf("该云盘需通过 OAuth 授权绑定")
	}
	return nil
}

// CheckCredentialOrBinding validates provider; credential or binding id required.
func CheckCredentialOrBinding(providerID, credential, bindingID, oauthBindingID string) error {
	if strings.TrimSpace(oauthBindingID) != "" || strings.TrimSpace(bindingID) != "" {
		if _, ok := GetProviderInfo(providerID); !ok {
			return errUnsupportedProvider(providerID)
		}
		return nil
	}
	return CheckCredential(providerID, credential)
}

// RegisterShareLinkProvider 注册分享链接类 Provider（无 CloudProvider 工厂）
func RegisterShareLinkProvider(info ProviderInfo, taskType int) {
	info.Enabled = true
	providerRegistry[info.ID] = &providerEntry{info: info, taskType: taskType}
}

func errUnsupportedProvider(name string) error {
	return &unsupportedProviderError{name: name}
}

type unsupportedProviderError struct {
	name string
}

func (e *unsupportedProviderError) Error() string {
	return "不支持的云盘类型: " + e.name
}
