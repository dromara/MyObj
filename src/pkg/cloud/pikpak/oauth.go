package pikpak

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
	pikpakOAuthAuthorizeURL = "https://passport.pikpak.com/authorize"
	pikpakOAuthTokenURL     = "https://passport.pikpak.com/v1/auth/signin"
	pikpakUserinfoURL       = "https://api-drive.mypikpak.com/v1/user/info"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	UserName     string    `json:"user_name"`
	UserID       string    `json:"user_id"`
}

type OAuthManager struct {
	config *OAuthConfig
}

func NewOAuthManager(clientID, clientSecret, redirectURI string) *OAuthManager {
	if clientID == "" {
		clientID = os.Getenv("PIKPAK_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("PIKPAK_CLIENT_SECRET")
	}
	if redirectURI == "" {
		redirectURI = os.Getenv("PIKPAK_REDIRECT_URI")
		if redirectURI == "" {
			redirectURI = "http://localhost:8080/api/cloud/pikpak/callback"
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

func (m *OAuthManager) GetAuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "user:read file:read file:write")
	params.Set("state", state)
	return pikpakOAuthAuthorizeURL + "?" + params.Encode()
}

func (m *OAuthManager) ExchangeCode(code string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", m.config.RedirectURI)

	resp, err := http.PostForm(pikpakOAuthTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("请求Token失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
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
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		TokenType:    tokenResp.TokenType,
	}, nil
}

func (m *OAuthManager) RefreshToken(refreshToken string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("client_secret", m.config.ClientSecret)
	params.Set("refresh_token", refreshToken)
	params.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(pikpakOAuthTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("刷新Token失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("刷新Token失败: status=%d", resp.StatusCode)
	}
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token失败: %w", err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("刷新Token失败: %s", tokenResp.Error)
	}
	return &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

type TokenStore struct {
	tokens map[string]*OAuthToken
	mu     sync.RWMutex
	dir    string
}

func NewTokenStore() *TokenStore {
	s := &TokenStore{
		tokens: make(map[string]*OAuthToken),
		dir:    "./data/pikpak_tokens",
	}
	os.MkdirAll(s.dir, 0755)
	s.loadAll()
	return s
}

func (s *TokenStore) Save(userID string, token *OAuthToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = token
	data, _ := json.MarshalIndent(token, "", "  ")
	return os.WriteFile(s.dir+"/"+userID+".json", data, 0644)
}

func (s *TokenStore) Get(userID string) (*OAuthToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tokens[userID]
	if !ok {
		return nil, fmt.Errorf("未登录PikPak")
	}
	return t, nil
}

func (s *TokenStore) Delete(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
	return os.Remove(s.dir + "/" + userID + ".json")
}

func (s *TokenStore) Has(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.tokens[userID]
	return ok
}

func (s *TokenStore) loadAll() {
	files, _ := os.ReadDir(s.dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := os.ReadFile(s.dir + "/" + f.Name())
		if err != nil {
			continue
		}
		var t OAuthToken
		if json.Unmarshal(data, &t) == nil {
			userID := f.Name()[:len(f.Name())-5]
			s.tokens[userID] = &t
		}
	}
}
