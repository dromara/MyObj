package middleware

import (
	"context"
	"math"
	"myobj/src/config"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/auth"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"myobj/src/pkg/util"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	cache          cache.Cache
	apiKeyRepo     repository.ApiKeyRepository
	userRepo       repository.UserRepository
	groupPowerRepo repository.GroupPowerRepository
	powerRepo      repository.PowerRepository
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(
	cache cache.Cache,
	apiKeyRepo repository.ApiKeyRepository,
	userRepo repository.UserRepository,
	groupPowerRepo repository.GroupPowerRepository,
	powerRepo repository.PowerRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		cache:          cache,
		apiKeyRepo:     apiKeyRepo,
		userRepo:       userRepo,
		groupPowerRepo: groupPowerRepo,
		powerRepo:      powerRepo,
	}
}

// Verify 认证验证中间件
func (m *AuthMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 尝试从Authorization头获取JWT Token
		authorization := c.Request.Header.Get("Authorization")

		if authorization != "" {
			// JWT认证流程
			if err := m.handleJWTAuth(c, authorization); err != nil {
				c.JSON(200, models.NewJsonResponse(401, err.Error(), nil))
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// 2. 如果没有JWT,检查是否启用了API Key
		if config.CONFIG.Auth.ApiKey {
			// API Key认证流程
			if err := m.handleAPIKeyAuth(c); err != nil {
				c.JSON(200, models.NewJsonResponse(401, err.Error(), nil))
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// 3. 没有任何认证信息
		c.JSON(200, models.NewJsonResponse(401, "未授权:缺少认证信息", nil))
		c.Abort()
	}
}

// handleJWTAuth 处理JWT认证
func (m *AuthMiddleware) handleJWTAuth(c *gin.Context, authorization string) error {
	// 解析Authorization头,支持 "Bearer {token}" 格式
	token := strings.TrimSpace(authorization)
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("未授权:Token为空")
	}
	get, err := m.cache.Get(token)
	if err != nil {
		return err
	}
	jwtToken := get.(string)
	// 解析JWT
	claims, err := auth.ParseToken(jwtToken)
	if err != nil {
		return fmt.Errorf("未授权:Token解析失败 - %v", err)
	}

	// 检查JWT是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return fmt.Errorf("未授权:Token已过期,请重新登录")
	}

	// 检查JWT剩余时间,如果不足5分钟则刷新
	if claims.ExpiresAt != nil {
		timeRemaining := time.Until(claims.ExpiresAt.Time)
		if timeRemaining > 0 && timeRemaining < 5*time.Minute {
			// 重新生成JWT
			newToken, err := auth.GenerateJWT(claims.UserID, claims.SessionID, claims.UserLogin)
			if err == nil {
				// 更新缓存中的token
				_ = m.cache.Set(token, newToken, 5*60)
			}
		}
	}
	id, err := m.powerRepo.GetByGroupID(context.Background(), claims.UserLogin.User.GroupID)
	if err != nil {
		return err
	}
	claims.UserLogin.Power = id
	// 将用户信息放入gin context
	c.Set("userLogin", claims.UserLogin)
	c.Set("userID", claims.UserID)
	return nil
}

// handleAPIKeyAuth 处理API Key认证
func (m *AuthMiddleware) handleAPIKeyAuth(c *gin.Context) error {
	// 获取API Key相关请求头
	apiKey := c.Request.Header.Get("X-API-Key")
	signature := c.Request.Header.Get("X-Signature")
	timestampStr := c.Request.Header.Get("X-Timestamp")
	nonce := c.Request.Header.Get("X-Nonce")

	// 检查必要参数
	if apiKey == "" || signature == "" || timestampStr == "" || nonce == "" {
		return fmt.Errorf("未授权:API Key认证参数不完整")
	}

	// 验证nonce不为空
	if strings.TrimSpace(nonce) == "" {
		return fmt.Errorf("未授权:nonce不能为空")
	}

	// 解析时间戳
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("未授权:时间戳格式错误")
	}

	// 验证时间戳(不能超过5分钟)
	requestTime := time.UnixMilli(timestamp)
	timeDiff := math.Abs(time.Since(requestTime).Minutes())
	if timeDiff > 5 {
		return fmt.Errorf("未授权:请求时间戳已过期")
	}

	// 查询API Key记录
	ctx := context.Background()
	apiKeyRecord, err := m.apiKeyRepo.GetByKey(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("未授权:API Key不存在")
	}

	// 检查API Key是否过期
	if !apiKeyRecord.ExpiresAt.IsZero() && time.Time(apiKeyRecord.ExpiresAt).Before(time.Now()) {
		return fmt.Errorf("未授权:API Key已过期")
	}

	// 使用私钥解密签名
	decryptedData, err := util.DecryptToString(apiKeyRecord.PrivateKey, signature)
	if err != nil {
		return fmt.Errorf("未授权:签名验证失败 - %v", err)
	}

	// 解析签名内容: apikey=""&timestamp=毫秒时间戳&nonce="随机字符串"
	parsedValues, err := url.ParseQuery(decryptedData)
	if err != nil {
		return fmt.Errorf("未授权:签名内容格式错误")
	}

	// 验证签名中的apikey
	signApiKey := parsedValues.Get("apikey")
	if signApiKey != apiKey {
		return fmt.Errorf("未授权:签名中的API Key不匹配")
	}

	// 验证签名中的时间戳
	signTimestamp := parsedValues.Get("timestamp")
	if signTimestamp != timestampStr {
		return fmt.Errorf("未授权:签名中的时间戳不匹配")
	}

	// 验证签名中的nonce
	signNonce := parsedValues.Get("nonce")
	if signNonce == "" || signNonce != nonce {
		return fmt.Errorf("未授权:签名中的nonce不匹配")
	}

	// 查询用户信息
	user, err := m.userRepo.GetByID(ctx, apiKeyRecord.UserID)
	if err != nil {
		return fmt.Errorf("未授权:用户不存在")
	}

	// 查询用户权限
	groupPowers, err := m.groupPowerRepo.GetByGroupID(ctx, user.GroupID)
	if err != nil {
		return fmt.Errorf("未授权:权限查询失败")
	}

	// 获取权限详情
	var powers []*models.Power
	for _, gp := range groupPowers {
		power, err := m.powerRepo.GetByID(ctx, gp.PowerID)
		if err == nil && power != nil {
			powers = append(powers, power)
		}
	}

	// 构造UserLoginResponse
	userLoginResp := response.UserLoginResponse{
		Token: "", // API Key认证不使用JWT Token
		User:  user,
		Power: powers,
	}

	// 将用户信息放入gin context
	c.Set("userLogin", userLoginResp)
	c.Set("userID", user.ID)

	return nil
}
