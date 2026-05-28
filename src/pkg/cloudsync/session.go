package cloudsync

import (
	"fmt"
	"myobj/src/pkg/cloudsync/internal"
	"strings"
	"sync"
	"time"
)

// CredentialSync 支持 refresh_token 轮换后回写凭据的 Provider
type CredentialSync interface {
	SetCredentialUpdateCallback(func(string))
	ExportCredential() string
}

// SessionOptions 创建 Provider 实例时的会话选项
type SessionOptions struct {
	OnCredentialUpdate func(newCredential string)
}

// OpenProvider 获取 Provider 并绑定凭据轮换回调
func OpenProvider(name, credential string, opts SessionOptions) (CloudProvider, error) {
	p, err := GetProvider(name, credential)
	if err != nil {
		return nil, err
	}
	if opts.OnCredentialUpdate != nil {
		if sync, ok := p.(CredentialSync); ok {
			sync.SetCredentialUpdateCallback(opts.OnCredentialUpdate)
		}
	}
	return p, nil
}

type cachedUserInfo struct {
	info      *CloudUserInfo
	expiresAt time.Time
}

var (
	validateCacheMu sync.RWMutex
	validateCache   = make(map[string]*cachedUserInfo)
	validateCacheTTL = 5 * time.Minute
)

// ValidateCached 校验 Provider 凭据，命中短期缓存则跳过重复请求
func ValidateCached(provider CloudProvider, providerID, credential string) (*CloudUserInfo, error) {
	key := providerID + ":" + internal.HashCredential(credential)

	validateCacheMu.RLock()
	if cached, ok := validateCache[key]; ok && time.Now().Before(cached.expiresAt) {
		info := *cached.info
		validateCacheMu.RUnlock()
		return &info, nil
	}
	validateCacheMu.RUnlock()

	info, err := provider.Validate()
	if err != nil {
		return nil, formatValidateError(providerID, err)
	}

	validateCacheMu.Lock()
	validateCache[key] = &cachedUserInfo{
		info:      info,
		expiresAt: time.Now().Add(validateCacheTTL),
	}
	validateCacheMu.Unlock()
	return info, nil
}

// InvalidateValidationCache 清除指定凭据的校验缓存（凭据更新后调用）
func InvalidateValidationCache(providerID, credential string) {
	key := providerID + ":" + internal.HashCredential(credential)
	validateCacheMu.Lock()
	delete(validateCache, key)
	validateCacheMu.Unlock()
}

func formatValidateError(providerID string, err error) error {
	msg := err.Error()
	lower := msg
	if containsAny(lower, "cookie", "Cookie", "401", "403", "unauthorized", "Unauthorized") {
		return fmt.Errorf("凭据无效或已过期，请重新获取: %w", err)
	}
	if info, ok := GetProviderInfo(providerID); ok && info.AuthType == AuthRefreshToken {
		if containsAny(lower, "refresh", "token", "Token") {
			return fmt.Errorf("refresh_token 无效或已过期，请重新授权: %w", err)
		}
	}
	return err
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if sub != "" && strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
