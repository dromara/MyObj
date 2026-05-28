package service

import (
	"context"
	"encoding/json"
	"fmt"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type oauthStatePayload struct {
	UserID   string `json:"user_id"`
	Provider string `json:"provider"`
}

// StartCloudOAuth 发起 OAuth 授权，返回跳转 URL
func (d *DownloadService) StartCloudOAuth(providerID, userID, baseURL string, cacheLocal cache.Cache) (*models.JsonResponse, error) {
	state := uuid.NewString()
	payload, _ := json.Marshal(oauthStatePayload{UserID: userID, Provider: providerID})
	if err := cacheLocal.Set("oauth_state:"+state, string(payload), 600); err != nil {
		return nil, fmt.Errorf("保存授权状态失败: %w", err)
	}

	authorizeURL, err := cloudsync.BuildOAuthAuthorizeURL(baseURL, providerID, state)
	if err != nil {
		return nil, err
	}

	return models.NewJsonResponse(200, "获取授权链接成功", map[string]interface{}{
		"authorize_url": authorizeURL,
		"state":         state,
	}), nil
}

// HandleCloudOAuthCallback 处理 OAuth 回调（骨架：换 Token 并保存绑定）
func (d *DownloadService) HandleCloudOAuthCallback(providerID, code, state string, baseURL string, cacheLocal cache.Cache) (*models.JsonResponse, error) {
	raw, err := cacheLocal.Get("oauth_state:" + state)
	if err != nil || raw == nil {
		return nil, fmt.Errorf("授权状态无效或已过期")
	}
	var payload oauthStatePayload
	switch v := raw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &payload); err != nil {
			return nil, fmt.Errorf("授权状态解析失败")
		}
	default:
		return nil, fmt.Errorf("授权状态格式错误")
	}
	if payload.Provider != providerID {
		return nil, fmt.Errorf("Provider 不匹配")
	}
	_ = cacheLocal.Delete("oauth_state:" + state)

	cfg, err := cloudsync.GetOAuthProvider(providerID)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, fmt.Errorf("OAuth 网盘 %s 尚未配置 ClientID/Secret，请在环境变量或配置文件中设置", providerID)
	}

	token, err := exchangeOAuthToken(cfg, code, baseURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	binding := &models.CloudOAuthBinding{
		ID:           uuid.NewString(),
		UserID:       payload.UserID,
		Provider:     providerID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		AccountName:  token.AccountName,
		CreatedAt:    custom_type.Now(),
		UpdatedAt:    custom_type.Now(),
	}
	if token.ExpiresIn > 0 {
		exp := custom_type.Now()
		binding.ExpiresAt = exp.Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	if err := d.factory.DB().WithContext(ctx).Save(binding).Error; err != nil {
		return nil, fmt.Errorf("保存 OAuth 绑定失败: %w", err)
	}

	return models.NewJsonResponse(200, "授权成功", map[string]interface{}{
		"provider":     providerID,
		"account_name": binding.AccountName,
	}), nil
}

// ListCloudOAuthBindings 列出用户 OAuth 绑定
func (d *DownloadService) ListCloudOAuthBindings(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	var bindings []models.CloudOAuthBinding
	if err := d.factory.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("查询绑定失败: %w", err)
	}
	list := make([]map[string]interface{}, 0, len(bindings))
	for _, b := range bindings {
		list = append(list, map[string]interface{}{
			"id":           b.ID,
			"provider":     b.Provider,
			"account_name": b.AccountName,
			"expires_at":   b.ExpiresAt,
			"created_at":   b.CreatedAt,
		})
	}
	return models.NewJsonResponse(200, "获取成功", list), nil
}

// DeleteCloudOAuthBinding 删除 OAuth 绑定
func (d *DownloadService) DeleteCloudOAuthBinding(userID, bindingID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	result := d.factory.DB().WithContext(ctx).
		Where("id = ? AND user_id = ?", bindingID, userID).
		Delete(&models.CloudOAuthBinding{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("绑定不存在")
	}
	return models.NewJsonResponse(200, "删除成功", nil), nil
}

type oauthTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	AccountName  string
}

func exchangeOAuthToken(cfg *cloudsync.OAuthProviderConfig, code, baseURL string) (*oauthTokenResult, error) {
	redirectURI := strings.TrimRight(baseURL, "/") + cfg.RedirectPath
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
		"client_id":    {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
	}

	req, err := http.NewRequest(http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Token 交换失败: %w", err)
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("解析 Token 响应失败: %w", err)
	}
	if errMsg, ok := body["error"].(string); ok && errMsg != "" {
		desc, _ := body["error_description"].(string)
		return nil, fmt.Errorf("OAuth 错误: %s %s", errMsg, desc)
	}

	result := &oauthTokenResult{}
	if v, ok := body["access_token"].(string); ok {
		result.AccessToken = v
	}
	if v, ok := body["refresh_token"].(string); ok {
		result.RefreshToken = v
	}
	if v, ok := body["expires_in"].(float64); ok {
		result.ExpiresIn = int64(v)
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("未获取到 access_token")
	}
	return result, nil
}
