package quark

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// 夸克网盘API地址
	quarkAPIBase = "https://drive.quark.cn"
)

// CookieConfig Cookie配置
type CookieConfig struct {
	UserID    string    `json:"user_id"`
	Cookie    string    `json:"cookie"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CookieAuthManager Cookie认证管理器
type CookieAuthManager struct {
	configs map[string]*CookieConfig
	mu      sync.RWMutex
	dataDir string
}

// NewCookieAuthManager 创建Cookie认证管理器
func NewCookieAuthManager() *CookieAuthManager {
	manager := &CookieAuthManager{
		configs: make(map[string]*CookieConfig),
		dataDir: "./data/quark_cookies",
	}
	
	// 创建数据目录
	os.MkdirAll(manager.dataDir, 0755)
	
	// 加载已保存的配置
	manager.loadConfigs()
	
	return manager
}

// SaveCookie 保存Cookie
func (m *CookieAuthManager) SaveCookie(userID, cookie, userName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	config := &CookieConfig{
		UserID:    userID,
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
	return m.saveConfig(userID, config)
}

// GetCookie 获取Cookie
func (m *CookieAuthManager) GetCookie(userID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, ok := m.configs[userID]
	if !ok {
		return "", fmt.Errorf("未配置夸克网盘Cookie")
	}
	
	return config.Cookie, nil
}

// GetConfig 获取配置
func (m *CookieAuthManager) GetConfig(userID string) (*CookieConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, ok := m.configs[userID]
	if !ok {
		return nil, fmt.Errorf("未配置夸克网盘Cookie")
	}
	
	return config, nil
}

// DeleteCookie 删除Cookie
func (m *CookieAuthManager) DeleteCookie(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.configs, userID)
	
	// 删除文件
	configFile := filepath.Join(m.dataDir, userID+".json")
	return os.Remove(configFile)
}

// ListConfigs 列出所有配置
func (m *CookieAuthManager) ListConfigs() []*CookieConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	configs := make([]*CookieConfig, 0, len(m.configs))
	for _, config := range m.configs {
		configs = append(configs, config)
	}
	
	return configs
}

// ValidateCookie 验证Cookie是否有效
func (m *CookieAuthManager) ValidateCookie(cookie string) (bool, string, error) {
	// 调用夸克API验证Cookie
	req, _ := http.NewRequest("GET", quarkAPIBase+"/1/clouddrive/file/sort?pdir_fid=0&_page=1&_size=10&_sort=file_type:asc,updated_at:desc&pr=ucpro&fr=pc", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch")
	req.Header.Set("Referer", "https://pan.quark.cn")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("验证Cookie失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("Cookie无效: status=%d", resp.StatusCode)
	}

	// 解析响应获取用户信息
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否成功
	if status, ok := result["status"].(float64); ok && status == 200 {
		// 获取用户名（从Cookie中提取）
		userName := "夸克用户"
		return true, userName, nil
	}

	return false, "", fmt.Errorf("Cookie无效")
}

// ListFiles 列出文件
func (m *CookieAuthManager) ListFiles(userID, dir string, page, size int) (map[string]interface{}, error) {
	cookie, err := m.GetCookie(userID)
	if err != nil {
		return nil, err
	}
	
	if dir == "" {
		dir = "0"
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	
	reqURL := fmt.Sprintf("%s/1/clouddrive/file/sort?pdir_fid=%s&_page=%d&_size=%d&_sort=file_type:asc,updated_at:desc&pr=ucpro&fr=pc",
		quarkAPIBase, dir, page, size)
	
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch")
	req.Header.Set("Referer", "https://pan.quark.cn")
	
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
func (m *CookieAuthManager) GetShareDetail(userID, shareID, pwd string) (map[string]interface{}, error) {
	cookie, err := m.GetCookie(userID)
	if err != nil {
		return nil, err
	}
	
	reqURL := fmt.Sprintf("%s/1/clouddrive/share/sharepage/detail?pr=ucpro&fr=pc", quarkAPIBase)
	
	req, _ := http.NewRequest("POST", reqURL, nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch")
	req.Header.Set("Referer", "https://pan.quark.cn")
	
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

// GetDownloadURL 获取下载链接
func (m *CookieAuthManager) GetDownloadURL(userID, fileID string) (string, error) {
	cookie, err := m.GetCookie(userID)
	if err != nil {
		return "", err
	}
	
	reqURL := fmt.Sprintf("%s/1/clouddrive/file/download?pr=ucpro&fr=pc", quarkAPIBase)
	
	req, _ := http.NewRequest("POST", reqURL, nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch")
	req.Header.Set("Referer", "https://pan.quark.cn")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取下载链接失败: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取下载链接失败: status=%d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析下载链接失败: %w", err)
	}
	
	// 提取下载链接
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if file, ok := data[0].(map[string]interface{}); ok {
			if url, ok := file["download_link"].(string); ok {
				return url, nil
			}
		}
	}
	
	return "", fmt.Errorf("未找到下载链接")
}

// saveConfig 保存配置到文件
func (m *CookieAuthManager) saveConfig(userID string, config *CookieConfig) error {
	configFile := filepath.Join(m.dataDir, userID+".json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	return os.WriteFile(configFile, data, 0644)
}

// loadConfigs 加载配置
func (m *CookieAuthManager) loadConfigs() {
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
		
		var config CookieConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}
		
		m.configs[userID] = &config
	}
}
