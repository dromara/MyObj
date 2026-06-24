package xunlei

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	// 迅雷网盘OAuth端点
	xunleiOAuthAuthorizeURL = "https://xluser-ssl.xunlei.com/v1/auth/signin"
	xunleiOAuthTokenURL     = "https://xluser-ssl.xunlei.com/v1/auth/token"
	xunleiUserinfoURL       = "https://xluser-ssl.xunlei.com/v1/user/me"
	xunleiDriveAPIURL       = "https://api-pan.xunlei.com/drive/v1"
	
	// 默认配置
	xunleiDefaultRedirectURI = "http://localhost:8080/api/cloud/xunlei/callback"
	
	// 迅雷应用配置
	xunleiAppID     = "40"
	xunleiAppKey    = "34a062aaa22f906fca4fefe9fb3a3021"
	xunleiClientID  = "40"
	xunleiClientSecret = "34a062aaa22f906fca4fefe9fb3a3021"
)

// OAuthConfig OAuth配置
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AppID        string
	AppKey       string
}

// OAuthToken OAuth Token
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
}

// OAuthManager OAuth管理器
type OAuthManager struct {
	config *OAuthConfig
}

// NewOAuthManager 创建OAuth管理器
func NewOAuthManager(clientID, clientSecret, redirectURI string) *OAuthManager {
	if clientID == "" {
		clientID = os.Getenv("XUNLEI_CLIENT_ID")
		if clientID == "" {
			clientID = xunleiClientID
		}
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("XUNLEI_CLIENT_SECRET")
		if clientSecret == "" {
			clientSecret = xunleiClientSecret
		}
	}
	if redirectURI == "" {
		redirectURI = os.Getenv("XUNLEI_REDIRECT_URI")
		if redirectURI == "" {
			redirectURI = xunleiDefaultRedirectURI
		}
	}
	
	return &OAuthManager{
		config: &OAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
			AppID:        xunleiAppID,
			AppKey:       xunleiAppKey,
		},
	}
}

// GetAuthorizeURL 获取授权URL
func (m *OAuthManager) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "drive:file:read drive:file:write drive:share:read")
	params.Set("state", state)
	
	return "https://xluser-ssl.xunlei.com/v1/auth/authorize?" + params.Encode()
}

// LoginWithPassword 使用用户名密码登录
func (m *OAuthManager) LoginWithPassword(username, password string) (*OAuthToken, error) {
	// 计算密码的MD5
	// passwordMD5 := md5.Sum([]byte(password))
	// passwordHex := hex.EncodeToString(passwordMD5[:])
	
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("username", username)
	params.Set("password", password)
	params.Set("grant_type", "password")
	
	resp, err := http.PostForm(xunleiOAuthTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登录失败: status=%d, body=%s", resp.StatusCode, string(body))
	}
	
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		UserID       string `json:"user_id"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("登录失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	
	// 获取用户信息
	userName := ""
	userInfo, err := m.GetUserInfo(tokenResp.AccessToken)
	if err == nil {
		if name, ok := userInfo["name"].(string); ok {
			userName = name
		}
	}
	
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		UserID:       tokenResp.UserID,
		UserName:     userName,
	}, nil
}

// ExchangeCode 用授权码换取Token
func (m *OAuthManager) ExchangeCode(code string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", m.config.RedirectURI)
	
	resp, err := http.PostForm(xunleiOAuthTokenURL, params)
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
		UserID       string `json:"user_id"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("授权失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	
	// 获取用户信息
	userName := ""
	userInfo, err := m.GetUserInfo(tokenResp.AccessToken)
	if err == nil {
		if name, ok := userInfo["name"].(string); ok {
			userName = name
		}
	}
	
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
		UserID:       tokenResp.UserID,
		UserName:     userName,
	}, nil
}

// RefreshToken 刷新Token
func (m *OAuthManager) RefreshToken(refreshToken string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("refresh_token", refreshToken)
	params.Set("grant_type", "refresh_token")
	
	resp, err := http.PostForm(xunleiOAuthTokenURL, params)
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
	}, nil
}

// GetUserInfo 获取用户信息
func (m *OAuthManager) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", xunleiUserinfoURL, nil)
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
	reqURL := xunleiDriveAPIURL + "/about"
	req, _ := http.NewRequest("GET", reqURL, nil)
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
func (m *OAuthManager) ListFiles(accessToken, parentID string, pageSize int) (map[string]interface{}, error) {
	if parentID == "" {
		parentID = "root"
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	
	params := url.Values{}
	params.Set("parent_id", parentID)
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	params.Set("order_by", "updated_at")
	params.Set("order_direction", "DESC")
	
	reqURL := xunleiDriveAPIURL + "/files?" + params.Encode()
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

// GetShareDetail 获取分享详情
func (m *OAuthManager) GetShareDetail(accessToken, shareID, pwd string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("share_id", shareID)
	if pwd != "" {
		params.Set("pwd", pwd)
	}
	
	reqURL := xunleiDriveAPIURL + "/share/link/info?" + params.Encode()
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
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
