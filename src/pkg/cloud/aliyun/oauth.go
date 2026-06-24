package aliyun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	aliyunOAuthAuthorizeURL = "https://openapi.aliyundrive.com/oauth/authorize"
	aliyunOAuthTokenURL     = "https://auth.aliyundrive.com/v2/account/token"
	aliyunUserinfoURL       = "https://openapi.aliyundrive.com/v2/user/get"
)

type AliyunOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type AliyunOAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	UserName     string    `json:"user_name"`
	UserID       string    `json:"user_id"`
}

type AliyunOAuthManager struct {
	config *AliyunOAuthConfig
}

func NewAliyunOAuthManager(clientID, clientSecret, redirectURI string) *AliyunOAuthManager {
	if clientID == "" {
		clientID = os.Getenv("ALIYUN_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("ALIYUN_CLIENT_SECRET")
	}
	if redirectURI == "" {
		redirectURI = os.Getenv("ALIYUN_REDIRECT_URI")
		if redirectURI == "" {
			redirectURI = "http://localhost:8080/api/cloud/aliyun/callback"
		}
	}
	return &AliyunOAuthManager{
		config: &AliyunOAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
		},
	}
}

func (m *AliyunOAuthManager) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "user:base,file:all:read,file:all:write")
	params.Set("state", state)
	return aliyunOAuthAuthorizeURL + "?" + params.Encode()
}

func (m *AliyunOAuthManager) ExchangeCode(code string) (*AliyunOAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", m.config.RedirectURI)

	resp, err := http.PostForm(aliyunOAuthTokenURL, params)
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
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("授权失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &AliyunOAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
	}, nil
}

func (m *AliyunOAuthManager) RefreshToken(refreshToken string) (*AliyunOAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("refresh_token", refreshToken)
	params.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(aliyunOAuthTokenURL, params)
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
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("刷新Token失败: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &AliyunOAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

// AliyunTokenStore 阿里云盘Token存储
type AliyunTokenStore struct {
	tokens map[string]*AliyunOAuthToken
	mu     sync.RWMutex
	dir    string
}

func NewAliyunTokenStore() *AliyunTokenStore {
	s := &AliyunTokenStore{
		tokens: make(map[string]*AliyunOAuthToken),
		dir:    "./data/aliyun_tokens",
	}
	os.MkdirAll(s.dir, 0755)
	s.loadAll()
	return s
}

func (s *AliyunTokenStore) Save(userID string, token *AliyunOAuthToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = token
	return s.saveToFile(userID, token)
}

func (s *AliyunTokenStore) Get(userID string) (*AliyunOAuthToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[userID]
	if !ok {
		return nil, fmt.Errorf("未登录阿里云盘")
	}
	return t, nil
}

func (s *AliyunTokenStore) Delete(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
	return os.Remove(s.dir + "/" + userID + ".json")
}

func (s *AliyunTokenStore) Has(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.tokens[userID]
	return ok
}

func (s *AliyunTokenStore) saveToFile(userID string, token *AliyunOAuthToken) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.dir+"/"+userID+".json", data, 0644)
}

func (s *AliyunTokenStore) loadAll() {
	files, _ := os.ReadDir(s.dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := os.ReadFile(s.dir + "/" + f.Name())
		if err != nil {
			continue
		}
		var t AliyunOAuthToken
		if json.Unmarshal(data, &t) == nil {
			userID := f.Name()[:len(f.Name())-5]
			s.tokens[userID] = &t
		}
	}
}
