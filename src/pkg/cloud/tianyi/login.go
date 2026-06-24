package tianyi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LoginConfig 登录配置
type LoginConfig struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Cookie    string    `json:"cookie"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginManager 登录管理器
type LoginManager struct {
	configs map[string]*LoginConfig
	mu      sync.RWMutex
	dataDir string
}

// NewLoginManager 创建登录管理器
func NewLoginManager() *LoginManager {
	manager := &LoginManager{
		configs: make(map[string]*LoginConfig),
		dataDir: "./data/tianyi_login",
	}
	
	// 创建数据目录
	os.MkdirAll(manager.dataDir, 0755)
	
	// 加载已保存的配置
	manager.loadConfigs()
	
	return manager
}

// Login 登录
func (m *LoginManager) Login(userID, username, password string) (string, error) {
	// 调用天翼云盘登录API
	cookie, userName, err := loginTianyi(username, password)
	if err != nil {
		return "", err
	}
	
	// 保存登录信息
	m.mu.Lock()
	defer m.mu.Unlock()
	
	config := &LoginConfig{
		UserID:    userID,
		Username:  username,
		Password:  password,
		Cookie:    cookie,
		UserName:  userName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 如果已存在，保留创建时间
	if existing, ok := m.configs[userID]; ok {
		config.CreatedAt = existing.CreatedAt
	}
	
	m.configs[userID] = config
	
	// 保存到文件
	if err := m.saveConfig(userID, config); err != nil {
		return "", fmt.Errorf("保存登录信息失败: %w", err)
	}
	
	return userName, nil
}

// GetCookie 获取Cookie
func (m *LoginManager) GetCookie(userID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, ok := m.configs[userID]
	if !ok {
		return "", fmt.Errorf("未登录天翼云盘")
	}
	
	return config.Cookie, nil
}

// GetConfig 获取配置
func (m *LoginManager) GetConfig(userID string) (*LoginConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, ok := m.configs[userID]
	if !ok {
		return nil, fmt.Errorf("未登录天翼云盘")
	}
	
	return config, nil
}

// Logout 登出
func (m *LoginManager) Logout(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.configs, userID)
	
	// 删除文件
	configFile := filepath.Join(m.dataDir, userID+".json")
	return os.Remove(configFile)
}

// IsLoggedIn 检查是否已登录
func (m *LoginManager) IsLoggedIn(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	_, ok := m.configs[userID]
	return ok
}

// ValidateCookie 验证Cookie是否有效
func (m *LoginManager) ValidateCookie(cookie string) (bool, error) {
	// 调用天翼云盘API验证Cookie
	req, _ := http.NewRequest("GET", "https://cloud.189.cn/api/portal/getUserSizeInfo.action", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("验证Cookie失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 如果返回200，说明Cookie有效
	return resp.StatusCode == http.StatusOK, nil
}

// GetUserSizeInfo 获取用户空间信息
func (m *LoginManager) GetUserSizeInfo(userID string) (map[string]interface{}, error) {
	cookie, err := m.GetCookie(userID)
	if err != nil {
		return nil, err
	}
	
	req, _ := http.NewRequest("GET", "https://cloud.189.cn/api/portal/getUserSizeInfo.action", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取用户空间信息失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户空间信息失败: status=%d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析用户空间信息失败: %w", err)
	}
	
	return result, nil
}

// loginTianyi 登录天翼云盘
func loginTianyi(username, password string) (string, string, error) {
	// 第一步：获取登录页面，获取lt和reqId
	loginURL := "https://cloud.189.cn/api/portal/loginUrl.action"
	req, _ := http.NewRequest("GET", loginURL, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("获取登录页面失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 获取Cookie
	cookies := resp.Cookies()
	cookieStr := ""
	for _, c := range cookies {
		cookieStr += c.Name + "=" + c.Value + "; "
	}
	
	// 第二步：提交登录
	loginPostURL := "https://cloud.189.cn/api/portal/loginSubmit.action"
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("lt", "")
	data.Set("reqId", "")
	data.Set("dynamicCheck", "FALSE")
	
	req, _ = http.NewRequest("POST", loginPostURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Cookie", cookieStr)
	
	resp, err = client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("解析登录响应失败: %w", err)
	}
	
	// 检查登录结果
	if code, ok := result["code"].(float64); ok && code != 0 {
		msg, _ := result["msg"].(string)
		return "", "", fmt.Errorf("登录失败: %s", msg)
	}
	
	// 获取用户信息
	userName := ""
	if data, ok := result["data"].(map[string]interface{}); ok {
		if name, ok := data["userName"].(string); ok {
			userName = name
		}
	}
	
	// 获取新的Cookie
	newCookies := resp.Cookies()
	for _, c := range newCookies {
		cookieStr += c.Name + "=" + c.Value + "; "
	}
	
	return cookieStr, userName, nil
}

// saveConfig 保存配置到文件
func (m *LoginManager) saveConfig(userID string, config *LoginConfig) error {
	configFile := filepath.Join(m.dataDir, userID+".json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	return os.WriteFile(configFile, data, 0644)
}

// loadConfigs 加载配置
func (m *LoginManager) loadConfigs() {
	files, err := os.ReadDir(m.dataDir)
	if err != nil {
		return
	}
	
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		userID := file.Name()[:len(file.Name())-5]
		configFile := filepath.Join(m.dataDir, file.Name())
		
		data, err := os.ReadFile(configFile)
		if err != nil {
			continue
		}
		
		var config LoginConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}
		
		m.configs[userID] = &config
	}
}
