package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"myobj/src/pkg/cloudsync/internal"
)

// oauthAccessSyncHelper writes access|refresh token back after refresh.
type oauthAccessSyncHelper struct {
	onUpdate func(string)
}

func (h *oauthAccessSyncHelper) SetCredentialUpdateCallback(fn func(string)) {
	h.onUpdate = fn
}

func (h *oauthAccessSyncHelper) notify(accessToken, refreshToken string) {
	if h.onUpdate == nil {
		return
	}
	cred := internal.FormatOAuthAccessCredential(accessToken, refreshToken)
	if cred != "" {
		h.onUpdate(cred)
	}
}

func refreshOAuthToken(tokenURL, clientID, clientSecret, refreshToken string) (accessToken, newRefresh string, expiresIn int64, err error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("请求token失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, err
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", 0, err
	}
	if errMsg, _ := parsed["error"].(string); errMsg != "" {
		desc, _ := parsed["error_description"].(string)
		return "", "", 0, fmt.Errorf("OAuth 刷新错误: %s %s", errMsg, desc)
	}
	accessToken, _ = parsed["access_token"].(string)
	newRefresh, _ = parsed["refresh_token"].(string)
	if v, ok := parsed["expires_in"].(float64); ok {
		expiresIn = int64(v)
	}
	if accessToken == "" {
		return "", "", 0, fmt.Errorf("token刷新失败: access_token为空")
	}
	if newRefresh == "" {
		newRefresh = refreshToken
	}
	return accessToken, newRefresh, expiresIn, nil
}

type oauthSession struct {
	accessToken  string
	refreshToken string
	mu           sync.RWMutex
	credSync     oauthAccessSyncHelper
	client       *http.Client
}

func newOAuthSession(credential string) *oauthSession {
	access, refresh := internal.ParseOAuthAccessCredential(credential)
	return &oauthSession{
		accessToken:  access,
		refreshToken: refresh,
		client:       internal.DefaultHTTPClient(),
	}
}

func (s *oauthSession) SetCredentialUpdateCallback(fn func(string)) {
	s.credSync.SetCredentialUpdateCallback(fn)
}

func (s *oauthSession) ExportCredential() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return internal.FormatOAuthAccessCredential(s.accessToken, s.refreshToken)
}

func (s *oauthSession) access() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

func (s *oauthSession) setTokens(access, refresh string) {
	s.mu.Lock()
	s.accessToken = access
	if refresh != "" {
		s.refreshToken = refresh
	}
	s.mu.Unlock()
	s.credSync.notify(access, refresh)
}

func (s *oauthSession) refresh(tokenURL, clientID, clientSecret string) error {
	s.mu.RLock()
	refreshToken := s.refreshToken
	s.mu.RUnlock()
	if refreshToken == "" {
		return fmt.Errorf("refresh_token 不能为空")
	}
	access, newRefresh, _, err := refreshOAuthToken(tokenURL, clientID, clientSecret, refreshToken)
	if err != nil {
		return err
	}
	s.setTokens(access, newRefresh)
	return nil
}

func (s *oauthSession) authorizedRequest(method, reqURL string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.access())
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return s.client.Do(req)
}
