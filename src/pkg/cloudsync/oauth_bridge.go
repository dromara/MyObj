package cloudsync

import "myobj/src/pkg/cloudsync/oauth"

// OAuthProviderConfig OAuth 网盘配置（兼容旧引用）
type OAuthProviderConfig = oauth.ProviderConfig

// OAuthProviderInfo OAuth Provider 对外信息
type OAuthProviderInfo = oauth.ProviderInfo

// ListOAuthProviders 返回 OAuth 网盘列表
func ListOAuthProviders(baseURL string) []OAuthProviderInfo {
	return oauth.ListProviders(baseURL)
}

// GetOAuthProvider 获取 OAuth 配置
func GetOAuthProvider(id string) (*OAuthProviderConfig, error) {
	return oauth.GetProvider(id)
}

// BuildOAuthAuthorizeURL 构建 OAuth 授权跳转 URL
func BuildOAuthAuthorizeURL(baseURL, providerID, state string) (string, error) {
	return oauth.BuildAuthorizeURL(baseURL, providerID, state)
}

// SetOAuthClientCredentials 运行时设置 OAuth 凭据
func SetOAuthClientCredentials(providerID, clientID, clientSecret string) {
	oauth.SetClientCredentials(providerID, clientID, clientSecret)
}
