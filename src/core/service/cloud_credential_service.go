package service

import (
	"context"
	"fmt"
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"

	"github.com/google/uuid"
)

type cloudCredentialContext struct {
	credential       string
	updateFn         func(string)
	oauthBindingID   string
	oauthAccessToken string
}

func (d *DownloadService) resolveCloudCredential(userID, provider, cookie, bindingID, oauthBindingID string) (*cloudCredentialContext, error) {
	ctx := context.Background()

	if oauthBindingID != "" {
		var binding models.CloudOAuthBinding
		if err := d.factory.DB().WithContext(ctx).
			Where("id = ? AND user_id = ? AND provider = ?", oauthBindingID, userID, provider).
			First(&binding).Error; err != nil {
			return nil, fmt.Errorf("OAuth 绑定不存在或无权访问")
		}
		cred := cloudsync.FormatOAuthAccessCredential(binding.AccessToken, binding.RefreshToken)
		bindingIDCopy := binding.ID
		userIDCopy := userID
		return &cloudCredentialContext{
			credential:       cred,
			oauthBindingID:   binding.ID,
			oauthAccessToken: binding.AccessToken,
			updateFn: func(newCred string) {
				access, refresh := cloudsync.ParseOAuthAccessCredential(newCred)
				if access == "" {
					return
				}
				_ = d.UpdateCloudOAuthRefreshToken(bindingIDCopy, userIDCopy, refresh, access, custom_type.JsonTime{})
			},
		}, nil
	}

	if bindingID != "" {
		var binding models.CloudCredentialBinding
		if err := d.factory.DB().WithContext(ctx).
			Where("id = ? AND user_id = ? AND provider = ?", bindingID, userID, provider).
			First(&binding).Error; err != nil {
			return nil, fmt.Errorf("凭据绑定不存在或无权访问")
		}
		bindingIDCopy := binding.ID
		return &cloudCredentialContext{
			credential: binding.Credential,
			updateFn: func(newCred string) {
				if newCred == "" || newCred == binding.Credential {
					return
				}
				_ = d.factory.DB().WithContext(context.Background()).
					Model(&models.CloudCredentialBinding{}).
					Where("id = ?", bindingIDCopy).
					Updates(map[string]interface{}{
						"credential": newCred,
						"updated_at": custom_type.Now(),
					}).Error
				cloudsync.InvalidateValidationCache(provider, newCred)
				cloudsync.InvalidateListCache(provider, newCred)
			},
		}, nil
	}

	if cookie == "" {
		info, ok := cloudsync.GetProviderInfo(provider)
		if ok && info.AuthType == cloudsync.AuthOAuth2 {
			return nil, fmt.Errorf("请选择 OAuth 授权账号")
		}
		return nil, fmt.Errorf("凭据不能为空")
	}
	return &cloudCredentialContext{credential: cookie}, nil
}

func (d *DownloadService) openCloudProvider(userID, provider, cookie, bindingID, oauthBindingID string) (cloudsync.CloudProvider, *cloudCredentialContext, error) {
	credCtx, err := d.resolveCloudCredential(userID, provider, cookie, bindingID, oauthBindingID)
	if err != nil {
		return nil, nil, err
	}
	p, err := cloudsync.OpenProvider(provider, credCtx.credential, cloudsync.SessionOptions{
		OnCredentialUpdate: credCtx.updateFn,
	})
	if err != nil {
		return nil, nil, err
	}
	return p, credCtx, nil
}

func (d *DownloadService) saveCloudCredentialBinding(userID, provider, credential, accountName string) (string, error) {
	ctx := context.Background()
	binding := &models.CloudCredentialBinding{
		ID:          uuid.NewString(),
		UserID:      userID,
		Provider:    provider,
		Credential:  credential,
		AccountName: accountName,
		CreatedAt:   custom_type.Now(),
		UpdatedAt:   custom_type.Now(),
	}
	if err := d.factory.DB().WithContext(ctx).Create(binding).Error; err != nil {
		return "", fmt.Errorf("保存凭据绑定失败: %w", err)
	}
	return binding.ID, nil
}

func (d *DownloadService) ListCloudCredentialBindings(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	var bindings []models.CloudCredentialBinding
	if err := d.factory.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("查询绑定失败: %w", err)
	}
	list := make([]map[string]interface{}, 0, len(bindings))
	for _, b := range bindings {
		list = append(list, map[string]interface{}{
			"id":           b.ID,
			"provider":     b.Provider,
			"account_name": b.AccountName,
			"updated_at":   b.UpdatedAt,
			"created_at":   b.CreatedAt,
		})
	}
	return models.NewJsonResponse(200, "获取成功", list), nil
}

func (d *DownloadService) DeleteCloudCredentialBinding(userID, bindingID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	result := d.factory.DB().WithContext(ctx).
		Where("id = ? AND user_id = ?", bindingID, userID).
		Delete(&models.CloudCredentialBinding{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("绑定不存在")
	}
	return models.NewJsonResponse(200, "删除成功", nil), nil
}

func (d *DownloadService) UpdateCloudOAuthRefreshToken(bindingID, userID, refreshToken, accessToken string, expiresAt custom_type.JsonTime) error {
	if refreshToken == "" && accessToken == "" {
		return nil
	}
	ctx := context.Background()
	updates := map[string]interface{}{
		"updated_at": custom_type.Now(),
	}
	if refreshToken != "" {
		updates["refresh_token"] = refreshToken
	}
	if accessToken != "" {
		updates["access_token"] = accessToken
	}
	if !expiresAt.IsZero() {
		updates["expires_at"] = expiresAt
	}
	result := d.factory.DB().WithContext(ctx).
		Model(&models.CloudOAuthBinding{}).
		Where("id = ? AND user_id = ?", bindingID, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("OAuth 绑定不存在")
	}
	return nil
}
