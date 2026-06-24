package baidu

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
	// 百度OAuth2端点
	oauthAuthorizeURL = "https://openapi.baidu.com/oauth/2.0/authorize"
	oauthTokenURL     = "https://openapi.baidu.com/oauth/2.0/token"
	
	// 默认客户端配置（从环境变量或配置文件读取）
	defaultClientID     = ""
	defaultClientSecret = ""
	defaultRedirectURI  = "http://localhost:8080/api/cloud/baidu/callback"
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
	Scope        string    `json:"scope"`
	UserID       string    `json:"userid"`
	UserName     string    `json:"username"`
}

// OAuthManager OAuth管理器
type OAuthManager struct {
	config *OAuthConfig
}

// NewOAuthManager 创建OAuth管理器
func NewOAuthManager(clientID, clientSecret, redirectURI string) *OAuthManager {
	// 优先使用传入的参数，否则从环境变量读取
	if clientID == "" {
		clientID = os.Getenv("BAIDU_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("BAIDU_CLIENT_SECRET")
	}
	if redirectURI == "" {
		redirectURI = os.Getenv("BAIDU_REDIRECT_URI")
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
// 用户访问此URL进行登录授权
func (m *OAuthManager) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "basic,netdisk")
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
		Scope        string `json:"scope"`
		UserID       string `json:"userid"`
		UserName     string `json:"username"`
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
		Scope:        tokenResp.Scope,
		UserID:       tokenResp.UserID,
		UserName:     tokenResp.UserName,
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
		Scope:        tokenResp.Scope,
	}, nil
}

// GetUserInfo 获取用户信息
func (m *OAuthManager) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", "https://pan.baidu.com/rest/2.0/xpan/nas?method=uinfo", nil)
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
	req, _ := http.NewRequest("GET", "https://pan.baidu.com/api/quota?checkfree=1&checkexpire=1", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
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
func (m *OAuthManager) ListFiles(accessToken, dir string, page, num int) (map[string]interface{}, error) {
	if dir == "" {
		dir = "/"
	}
	if page <= 0 {
		page = 1
	}
	if num <= 0 {
		num = 100
	}
	
	params := url.Values{}
	params.Set("dir", dir)
	params.Set("order", "time")
	params.Set("start", fmt.Sprintf("%d", (page-1)*num))
	params.Set("limit", fmt.Sprintf("%d", num))
	
	reqURL := "https://pan.baidu.com/rest/2.0/xpan/file?" + params.Encode()
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
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

// ShareCreate 创建分享链接
func (m *OAuthManager) ShareCreate(accessToken string, fidList []int64, pwd string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("schannel", "4")
	params.Set("channel", "chunlei")
	
	fidListJSON, _ := json.Marshal(fidList)
	params.Set("fid_list", string(fidListJSON))
	
	if pwd != "" {
		params.Set("pwd", pwd)
	}
	
	reqURL := "https://pan.baidu.com/rest/2.0/xpan/share?" + params.Encode()
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("创建分享失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("创建分享失败: status=%d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析分享结果失败: %w", err)
	}
	
	return result, nil
}
