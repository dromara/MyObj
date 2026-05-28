package oauth

import (
	"fmt"
	"myobj/src/config"
	"net/url"
	"strings"
	"sync"
)

// ProviderConfig OAuth 网盘配置
type ProviderConfig struct {
	ID           string
	Name         string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	ClientID     string
	ClientSecret string
	RedirectPath string
	Enabled      bool
}

var (
	providers   = make(map[string]*ProviderConfig)
	providersMu sync.RWMutex
)

func init() {
	registerDefaults()
}

func registerDefaults() {
	configs := []*ProviderConfig{
		{
			ID: "onedrive", Name: "Microsoft OneDrive",
			AuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			Scopes:       []string{"offline_access", "Files.Read"},
			RedirectPath: "/api/download/cloud/oauth/callback/onedrive",
		},
		{
			ID: "google", Name: "Google Drive",
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
			RedirectPath: "/api/download/cloud/oauth/callback/google",
		},
		{
			ID: "dropbox", Name: "Dropbox",
			AuthURL:      "https://www.dropbox.com/oauth2/authorize",
			TokenURL:     "https://api.dropboxapi.com/oauth2/token",
			RedirectPath: "/api/download/cloud/oauth/callback/dropbox",
		},
	}
	for _, cfg := range configs {
		cfg.Enabled = cfg.ClientID != "" && cfg.ClientSecret != ""
		providers[cfg.ID] = cfg
	}
}

// LoadFromConfig 从 config.toml [cloud.oauth] 加载 OAuth 凭据
func LoadFromConfig(oauthCfg config.CloudOAuth) {
	applyCredentials("onedrive", oauthCfg.OnedriveClientID, oauthCfg.OnedriveClientSecret)
	applyCredentials("google", oauthCfg.GoogleClientID, oauthCfg.GoogleClientSecret)
	applyCredentials("dropbox", oauthCfg.DropboxClientID, oauthCfg.DropboxClientSecret)
}

func applyCredentials(id, clientID, clientSecret string) {
	providersMu.Lock()
	defer providersMu.Unlock()
	cfg, ok := providers[id]
	if !ok {
		return
	}
	if clientID != "" {
		cfg.ClientID = clientID
	}
	if clientSecret != "" {
		cfg.ClientSecret = clientSecret
	}
	cfg.Enabled = cfg.ClientID != "" && cfg.ClientSecret != ""
}

// ProviderInfo OAuth Provider 对外信息
type ProviderInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Scopes       []string `json:"scopes,omitempty"`
	Enabled      bool     `json:"enabled"`
	AuthorizeURL string   `json:"authorize_url,omitempty"`
}

// ListProviders 返回 OAuth 网盘列表
func ListProviders(baseURL string) []ProviderInfo {
	providersMu.RLock()
	defer providersMu.RUnlock()

	list := make([]ProviderInfo, 0, len(providers))
	for _, cfg := range providers {
		info := ProviderInfo{
			ID:      cfg.ID,
			Name:    cfg.Name,
			Scopes:  cfg.Scopes,
			Enabled: cfg.Enabled,
		}
		list = append(list, info)
	}
	return list
}

// GetProvider 获取 OAuth 配置
func GetProvider(id string) (*ProviderConfig, error) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	cfg, ok := providers[id]
	if !ok {
		return nil, fmt.Errorf("不支持的 OAuth 网盘: %s", id)
	}
	return cfg, nil
}

// BuildAuthorizeURL 构建 OAuth 授权跳转 URL
func BuildAuthorizeURL(baseURL, providerID, state string) (string, error) {
	cfg, err := GetProvider(providerID)
	if err != nil {
		return "", err
	}
	if !cfg.Enabled {
		return "", fmt.Errorf("OAuth 网盘 %s 尚未配置 ClientID/Secret，请在 config.toml [cloud.oauth] 中设置", providerID)
	}
	redirectURI := strings.TrimRight(baseURL, "/") + cfg.RedirectPath
	q := url.Values{
		"client_id":     {cfg.ClientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	if len(cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	if providerID == "google" {
		q.Set("access_type", "offline")
		q.Set("prompt", "consent")
	}
	return cfg.AuthURL + "?" + q.Encode(), nil
}

// SetClientCredentials 运行时设置 OAuth 凭据
func SetClientCredentials(providerID, clientID, clientSecret string) {
	applyCredentials(providerID, clientID, clientSecret)
}
