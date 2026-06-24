package p115

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type LoginConfig struct {
	UserID    string    `json:"user_id"`
	Cookie    string    `json:"cookie"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginManager struct {
	configs map[string]*LoginConfig
	mu      sync.RWMutex
	dataDir string
}

func NewLoginManager() *LoginManager {
	m := &LoginManager{
		configs: make(map[string]*LoginConfig),
		dataDir: "./data/115_cookies",
	}
	os.MkdirAll(m.dataDir, 0755)
	m.loadConfigs()
	return m
}

func (m *LoginManager) SaveCookie(userID, cookie, userName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	config := &LoginConfig{
		UserID:    userID,
		Cookie:    cookie,
		UserName:  userName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if existing, ok := m.configs[userID]; ok {
		config.CreatedAt = existing.CreatedAt
	}
	m.configs[userID] = config
	return m.saveConfig(userID, config)
}

func (m *LoginManager) GetCookie(userID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.configs[userID]
	if !ok {
		return "", fmt.Errorf("未配置115网盘Cookie")
	}
	return c.Cookie, nil
}

func (m *LoginManager) DeleteCookie(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, userID)
	return os.Remove(m.dataDir + "/" + userID + ".json")
}

func (m *LoginManager) IsLoggedIn(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.configs[userID]
	return ok
}

func (m *LoginManager) ValidateCookie(cookie string) (bool, string, error) {
	req, _ := http.NewRequest("GET", "https://webapi.115.com/user/info", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return false, "", fmt.Errorf("status=%d", resp.StatusCode)
	}
	var result struct {
		UserID   string `json:"user_id"`
		NickName string `json:"nick_name"`
		State    int    `json:"state"`
	}
	if json.Unmarshal(body, &result) == nil && result.State == 0 {
		return true, result.NickName, nil
	}
	return false, "", nil
}

func (m *LoginManager) saveConfig(userID string, config *LoginConfig) error {
	data, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(m.dataDir+"/"+userID+".json", data, 0644)
}

func (m *LoginManager) loadConfigs() {
	files, _ := os.ReadDir(m.dataDir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := os.ReadFile(m.dataDir + "/" + f.Name())
		if err != nil {
			continue
		}
		var c LoginConfig
		if json.Unmarshal(data, &c) == nil {
			userID := f.Name()[:len(f.Name())-5]
			m.configs[userID] = &c
		}
	}
}
