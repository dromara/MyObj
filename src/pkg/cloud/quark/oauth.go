package quark

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// 夸克网盘OAuth端点
	oauthAuthorizeURL = "https://open-api-drive.quark.cn/oauth/authorize"
	oauthTokenURL     = "https://open-api-drive.quark.cn/oauth/token"
	oauthUserinfoURL  = "https://open-api-drive.quark.cn/user/info"
	oauthQuotaURL     = "https://drive-pc.quark.cn/1/clouddrive/capacity"
	
	// 默认配置
	defaultRedirectURI = "http://localhost:8080/api/cloud/quark/callback"
)

// OAuthConfig OAuth配置
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// OAuthToken OAuth Token
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
}

// OAuthManager OAuth管理器
type OAuthManager struct {
	config *OAuthConfig
}

// NewOAuthManager 创建OAuth管理器
func NewOAuthManager(clientID, clientSecret, redirectURI string) *OAuthManager {
	if clientID == "" {
		clientID = os.Getenv("QUARK_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("QUARK_CLIENT_SECRET")
	}
	if redirectURI == "" {
		redirectURI = os.Getenv("QUARK_REDIRECT_URI")
		if redirectURI == "" {
			redirectURI = defaultRedirectURI
		}
	}
	
	return &OAuthManager{
		config: &OAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
		},
	}
}

// GetAuthorizeURL 获取授权URL
func (m *OAuthManager) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "drive:file:read drive:file:write drive:share:read drive:share:write")
	params.Set("state", state)
	
	return oauthAuthorizeURL + "?" + params.Encode()
}

// ExchangeCode 用授权码换取Token
func (m *OAuthManager) ExchangeCode(code string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", m.config.RedirectURI)
	
	resp, err := http.PostForm(oauthTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("请求Token失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取Token失败: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("授权失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
	}, nil
}

// RefreshToken 刷新Token
func (m *OAuthManager) RefreshToken(refreshToken string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("refresh_token", refreshToken)
	params.Set("grant_type", "refresh_token")
	
	resp, err := http.PostForm(oauthTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("刷新Token失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("刷新Token失败: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("刷新Token失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
	}, nil
}

// GetUserInfo 获取用户信息
func (m *OAuthManager) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", oauthUserinfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户信息失败: status=%d", resp.StatusCode)
	}
	
	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}
	
	return userInfo, nil
}

// GetQuota 获取网盘配额
func (m *OAuthManager) GetQuota(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", oauthQuotaURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取配额失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取配额失败: status=%d", resp.StatusCode)
	}
	
	var quota map[string]interface{}
	if err := json.Unmarshal(body, &quota); err != nil {
		return nil, fmt.Errorf("解析配额失败: %w", err)
	}
	
	return quota, nil
}

// ListFiles 列出文件
func (m *OAuthManager) ListFiles(accessToken, dir string, page, size int) (map[string]interface{}, error) {
	if dir == "" {
		dir = "0"
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	
	params := url.Values{}
	params.Set("pdir_fid", dir)
	params.Set("_page", fmt.Sprintf("%d", page))
	params.Set("_size", fmt.Sprintf("%d", size))
	params.Set("_sort", "file_type:asc,updated_at:desc")
	
	reqURL := "https://drive-pc.quark.cn/1/clouddrive/file/sort?" + params.Encode()
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("列出文件失败: status=%d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析文件列表失败: %w", err)
	}
	
	return result, nil
}

// GetShareDetail 获取分享详情
func (m *OAuthManager) GetShareDetail(accessToken, shareID, pwd string) (map[string]interface{}, error) {
	reqBody := fmt.Sprintf(`{"share_id":"%s","password":"%s","page":1,"size":100}`, shareID, pwd)
	
	req, _ := http.NewRequest(
		"POST",
		"https://drive-pc.quark.cn/1/clouddrive/share/sharepage/detail",
		strings.NewReader(reqBody),
	)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取分享详情失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取分享详情失败: status=%d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析分享详情失败: %w", err)
	}
	
	return result, nil
}
