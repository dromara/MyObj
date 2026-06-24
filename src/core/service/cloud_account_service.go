package service

import (
	"context"
	"fmt"
	"time"

	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cloud/aliyun"
	"myobj/src/pkg/cloud/pikpak"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
)

type CloudAccountService struct {
	factory *impl.RepositoryFactory
}

func NewCloudAccountService(factory *impl.RepositoryFactory) *CloudAccountService {
	// 自动建表
	db := factory.DB()
	if db != nil {
		db.AutoMigrate(&models.CloudAccount{})
	}
	return &CloudAccountService{factory: factory}
}

func (s *CloudAccountService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// ListAccounts 获取用户的云盘账号列表
func (s *CloudAccountService) ListAccounts(ctx context.Context, userID string) ([]models.CloudAccount, error) {
	var accounts []models.CloudAccount
	err := s.factory.DB().Where("user_id = ?", userID).Order("provider, id").Find(&accounts).Error
	return accounts, err
}

// GetAccount 获取单个账号
func (s *CloudAccountService) GetAccount(ctx context.Context, userID, provider string) (*models.CloudAccount, error) {
	var account models.CloudAccount
	err := s.factory.DB().Where("user_id = ? AND provider = ?", userID, provider).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// SaveOAuthToken 保存OAuth Token
func (s *CloudAccountService) SaveOAuthToken(ctx context.Context, userID, provider, accountName, accessToken, refreshToken string, expiresIn int) error {
	now := custom_type.Now()
	expiresAt := custom_type.JsonTime(time.Now().Add(time.Duration(expiresIn) * time.Second))

	account := &models.CloudAccount{
		UserID:       userID,
		Provider:     provider,
		AccountName:  accountName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    &expiresAt,
		Status:       models.CloudAccountStatusValid,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 尝试更新已有记录
	var existing models.CloudAccount
	err := s.factory.DB().Where("user_id = ? AND provider = ?", userID, provider).First(&existing).Error
	if err == nil {
		existing.AccessToken = accessToken
		existing.RefreshToken = refreshToken
		existing.ExpiresAt = &expiresAt
		existing.AccountName = accountName
		existing.Status = models.CloudAccountStatusValid
		existing.UpdatedAt = now
		return s.factory.DB().Save(&existing).Error
	}

	return s.factory.DB().Create(account).Error
}

// SaveCookie 保存Cookie认证
func (s *CloudAccountService) SaveCookie(ctx context.Context, userID, provider, accountName, cookie string) error {
	now := custom_type.Now()
	account := &models.CloudAccount{
		UserID:      userID,
		Provider:    provider,
		AccountName: accountName,
		Cookie:      cookie,
		Status:      models.CloudAccountStatusValid,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	var existing models.CloudAccount
	err := s.factory.DB().Where("user_id = ? AND provider = ?", userID, provider).First(&existing).Error
	if err == nil {
		existing.Cookie = cookie
		existing.AccountName = accountName
		existing.Status = models.CloudAccountStatusValid
		existing.UpdatedAt = now
		return s.factory.DB().Save(&existing).Error
	}

	return s.factory.DB().Create(account).Error
}

// DeleteAccount 删除账号
func (s *CloudAccountService) DeleteAccount(ctx context.Context, userID, provider string) error {
	return s.factory.DB().Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.CloudAccount{}).Error
}

// GetValidToken 获取有效的Token（自动刷新）
func (s *CloudAccountService) GetValidToken(ctx context.Context, userID, provider string) (string, error) {
	account, err := s.GetAccount(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("未绑定%s账号", provider)
	}

	if account.Status != models.CloudAccountStatusValid {
		return "", fmt.Errorf("%s账号状态异常", provider)
	}

	// 检查是否过期
	if account.ExpiresAt != nil && time.Now().After(account.ExpiresAt.ToTime()) {
		// 尝试刷新
		if account.RefreshToken != "" {
			newToken, refreshErr := s.refreshProviderToken(provider, account.RefreshToken)
			if refreshErr == nil {
				s.SaveOAuthToken(ctx, userID, provider, account.AccountName, newToken.AccessToken, newToken.RefreshToken, newToken.ExpiresIn)
				return newToken.AccessToken, nil
			}
		}
		// 刷新失败，标记为过期
		s.factory.DB().Model(account).Update("status", models.CloudAccountStatusExpired)
		return "", fmt.Errorf("%s账号已过期，请重新登录", provider)
	}

	if account.AccessToken != "" {
		return account.AccessToken, nil
	}
	if account.Cookie != "" {
		return account.Cookie, nil
	}

	return "", fmt.Errorf("%s账号凭证为空", provider)
}

type refreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (s *CloudAccountService) refreshProviderToken(provider, refreshToken string) (*refreshTokenResult, error) {
	switch provider {
	case "aliyun":
		m := aliyun.NewAliyunOAuthManager("", "", "")
		token, err := m.RefreshToken(refreshToken)
		if err != nil {
			return nil, err
		}
		return &refreshTokenResult{token.AccessToken, token.RefreshToken, token.ExpiresIn}, nil
	case "pikpak":
		m := pikpak.NewOAuthManager("", "", "")
		token, err := m.RefreshToken(refreshToken)
		if err != nil {
			return nil, err
		}
		return &refreshTokenResult{token.AccessToken, token.RefreshToken, token.ExpiresIn}, nil
	default:
		return nil, fmt.Errorf("provider %s does not support token refresh", provider)
	}
}

// CheckAllStatus 检查所有账号状态
func (s *CloudAccountService) CheckAllStatus(ctx context.Context, userID string) map[string]interface{} {
	accounts, _ := s.ListAccounts(ctx, userID)
	result := make(map[string]interface{})

	providers := []string{"aliyun", "baidu", "xunlei", "quark", "115", "tianyi", "uc", "caiyun", "wopan", "pikpak"}
	for _, p := range providers {
		result[p] = map[string]interface{}{
			"connected": false,
			"status":    "disconnected",
		}
	}

	for _, acc := range accounts {
		status := "connected"
		if acc.Status == models.CloudAccountStatusExpired {
			status = "expired"
		} else if acc.Status == models.CloudAccountStatusInvalid {
			status = "invalid"
		}
		result[acc.Provider] = map[string]interface{}{
			"connected":   acc.Status == models.CloudAccountStatusValid,
			"status":      status,
			"account_name": acc.AccountName,
			"updated_at":  acc.UpdatedAt,
		}
	}

	return result
}
