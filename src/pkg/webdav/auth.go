package webdav

import (
	"context"
	"fmt"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"time"
)

// Authenticator WebDAV 认证器
type Authenticator struct {
	apiKeyRepo repository.ApiKeyRepository
	userRepo   repository.UserRepository
	powerRepo  repository.PowerRepository
}

// NewAuthenticator 创建 WebDAV 认证器
func NewAuthenticator(
	apiKeyRepo repository.ApiKeyRepository,
	userRepo repository.UserRepository,
	powerRepo repository.PowerRepository,
) *Authenticator {
	return &Authenticator{
		apiKeyRepo: apiKeyRepo,
		userRepo:   userRepo,
		powerRepo:  powerRepo,
	}
}

// Authenticate WebDAV 认证（使用 API Key）
// username: 用户名
// password: API Key（直接使用，无需签名）
func (a *Authenticator) Authenticate(username, password string) (*models.UserInfo, error) {
	ctx := context.Background()

	// 1. 查询用户
	user, err := a.userRepo.GetByUserName(ctx, username)
	if err != nil {
		logger.LOG.Warn("WebDAV 认证失败：用户不存在", "username", username)
		return nil, fmt.Errorf("用户不存在")
	}

	// 2. 验证 API Key
	apiKeyRecord, err := a.apiKeyRepo.GetByKey(ctx, password)
	if err != nil {
		logger.LOG.Warn("WebDAV 认证失败：API Key 无效", "username", username)
		return nil, fmt.Errorf("API Key 无效")
	}

	// 3. 验证 API Key 是否属于该用户
	if apiKeyRecord.UserID != user.ID {
		logger.LOG.Warn("WebDAV 认证失败：API Key 与用户不匹配",
			"username", username,
			"api_key_user_id", apiKeyRecord.UserID,
			"request_user_id", user.ID,
		)
		return nil, fmt.Errorf("API Key 与用户不匹配")
	}

	// 4. 检查 API Key 是否过期
	if !apiKeyRecord.ExpiresAt.IsZero() && time.Time(apiKeyRecord.ExpiresAt).Before(time.Now()) {
		logger.LOG.Warn("WebDAV 认证失败：API Key 已过期",
			"username", username,
			"expires_at", apiKeyRecord.ExpiresAt,
		)
		return nil, fmt.Errorf("API Key 已过期")
	}

	// 5. 检查用户状态
	if user.State == 1 {
		logger.LOG.Warn("WebDAV 认证失败：用户已被禁用", "username", username, "user_id", user.ID)
		return nil, fmt.Errorf("用户已被禁用")
	}

	logger.LOG.Info("WebDAV 认证成功", "username", username, "user_id", user.ID)
	return user, nil
}

// CheckPermission 检查用户是否有 WebDAV 访问权限
func (a *Authenticator) CheckPermission(userID string, groupID int, permission string) (bool, error) {
	ctx := context.Background()

	// 查询用户的所有权限
	powers, err := a.powerRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		logger.LOG.Error("查询用户权限失败", "user_id", userID, "group_id", groupID, "error", err)
		return false, err
	}

	// 检查是否有指定权限
	for _, power := range powers {
		if power.Characteristic == permission {
			return true, nil
		}
	}

	logger.LOG.Warn("用户无 WebDAV 权限",
		"user_id", userID,
		"group_id", groupID,
		"required_permission", permission,
	)
	return false, nil
}
