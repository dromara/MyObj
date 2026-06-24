package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"myobj/src/core/service"
	"myobj/src/pkg/cloud/baidu"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// OAuthState 临时存储OAuth状态
type OAuthState struct {
	State     string    `json:"state"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// BaiduToken 百度Token存储
type BaiduToken struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserName     string    `json:"user_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CloudOAuthHandler 云盘OAuth处理器
type CloudOAuthHandler struct {
	oauthManager  *baidu.OAuthManager
	accountService *service.CloudAccountService
	states        map[string]*OAuthState
	tokens        map[string]*BaiduToken
	mu            sync.RWMutex
	tokenDir      string
}

// NewCloudOAuthHandler 创建云盘OAuth处理器
func NewCloudOAuthHandler(accountService *service.CloudAccountService) *CloudOAuthHandler {
	handler := &CloudOAuthHandler{
		oauthManager:  baidu.NewOAuthManager("", "", ""),
		accountService: accountService,
		states:        make(map[string]*OAuthState),
		tokens:        make(map[string]*BaiduToken),
		tokenDir:      "./data/baidu_tokens",
	}
	
	// 创建token存储目录
	os.MkdirAll(handler.tokenDir, 0755)
	
	// 加载已保存的token
	handler.loadTokens()
	
	return handler
}

// Router 注册OAuth路由
func (h *CloudOAuthHandler) Router(c *gin.RouterGroup) {
	oauth := c.Group("/cloud/baidu")
	{
		// 获取授权URL
		oauth.GET("/auth", h.GetAuthURL)
		// OAuth回调
		oauth.GET("/callback", h.HandleCallback)
		// 获取登录状态
		oauth.GET("/status", h.GetLoginStatus)
		// 登出
		oauth.POST("/logout", h.Logout)
		// 获取用户信息
		oauth.GET("/userinfo", h.GetUserInfo)
	}
}

// GetAuthURL 获取百度授权URL
func (h *CloudOAuthHandler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	// 生成随机state
	state := generateRandomState()
	
	// 保存state
	h.mu.Lock()
	h.states[state] = &OAuthState{
		State:     state,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	h.mu.Unlock()
	
	// 获取授权URL
	authURL := h.oauthManager.GetAuthorizeURL(state)
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"auth_url": authURL,
		"state":    state,
	}))
}

// HandleCallback 处理OAuth回调
func (h *CloudOAuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" || state == "" {
		c.JSON(400, models.NewJsonResponse(400, "缺少必要参数", nil))
		return
	}
	
	// 验证state
	h.mu.RLock()
	oauthState, exists := h.states[state]
	h.mu.RUnlock()
	
	if !exists {
		c.JSON(400, models.NewJsonResponse(400, "无效的state参数", nil))
		return
	}
	
	// 删除已使用的state
	h.mu.Lock()
	delete(h.states, state)
	h.mu.Unlock()
	
	// 用授权码换取Token
	token, err := h.oauthManager.ExchangeCode(code)
	if err != nil {
		logger.LOG.Error("获取百度Token失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取Token失败: "+err.Error(), nil))
		return
	}
	
	// 保存Token
	baiduToken := &BaiduToken{
		UserID:       oauthState.UserID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
		UserName:     token.UserName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	h.mu.Lock()
	h.tokens[oauthState.UserID] = baiduToken
	h.mu.Unlock()
	
	// 保存到文件
	h.saveToken(oauthState.UserID, baiduToken)

	// 保存到数据库
	if h.accountService != nil {
		expiresIn := int(time.Until(token.ExpiresAt).Seconds())
		if err := h.accountService.SaveOAuthToken(nil, oauthState.UserID, "baidu", token.UserName, token.AccessToken, token.RefreshToken, expiresIn); err != nil {
			logger.LOG.Error("保存百度Token到数据库失败", "error", err)
		}
	}

	// 返回成功页面
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>百度网盘授权成功</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #4CAF50; font-size: 24px; margin-bottom: 20px; }
        .info { color: #666; margin-bottom: 10px; }
        .btn { background: #4CAF50; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; }
        .btn:hover { background: #45a049; }
    </style>
</head>
<body>
    <div class="success">✅ 百度网盘授权成功！</div>
    <div class="info">用户: %s</div>
    <div class="info">Token有效期: %s</div>
    <br>
    <button class="btn" onclick="window.close()">关闭窗口</button>
    <script>
        // 通知父窗口授权成功
        if (window.opener) {
            window.opener.postMessage({type: 'baidu_auth_success', user: '%s'}, '*');
        }
        // 3秒后自动关闭
        setTimeout(() => window.close(), 3000);
    </script>
</body>
</html>
`, token.UserName, token.ExpiresAt.Format("2006-01-02 15:04:05"), token.UserName)
}

// GetLoginStatus 获取登录状态
func (h *CloudOAuthHandler) GetLoginStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"logged_in": false,
		}))
		return
	}
	
	h.mu.RLock()
	token, exists := h.tokens[userID]
	h.mu.RUnlock()
	
	if !exists {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"logged_in": false,
		}))
		return
	}
	
	// 检查Token是否过期
	if time.Now().After(token.ExpiresAt) {
		// 尝试刷新Token
		newToken, err := h.oauthManager.RefreshToken(token.RefreshToken)
		if err != nil {
			logger.LOG.Error("刷新百度Token失败", "error", err)
			c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
				"logged_in":  false,
				"error":      "token_expired",
			}))
			return
		}
		
		// 更新Token
		token.AccessToken = newToken.AccessToken
		token.RefreshToken = newToken.RefreshToken
		token.ExpiresAt = newToken.ExpiresAt
		token.UpdatedAt = time.Now()
		
		h.mu.Lock()
		h.tokens[userID] = token
		h.mu.Unlock()
		
		h.saveToken(userID, token)
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"logged_in":  true,
		"user_name":  token.UserName,
		"expires_at": token.ExpiresAt,
	}))
}

// Logout 登出
func (h *CloudOAuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	h.mu.Lock()
	delete(h.tokens, userID)
	h.mu.Unlock()
	
	// 删除保存的文件
	tokenFile := filepath.Join(h.tokenDir, userID+".json")
	os.Remove(tokenFile)

	// 从数据库删除
	if h.accountService != nil {
		h.accountService.DeleteAccount(nil, userID, "baidu")
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// GetUserInfo 获取用户信息
func (h *CloudOAuthHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	h.mu.RLock()
	token, exists := h.tokens[userID]
	h.mu.RUnlock()
	
	if !exists {
		c.JSON(400, models.NewJsonResponse(400, "未登录百度网盘", nil))
		return
	}
	
	// 获取用户信息
	userInfo, err := h.oauthManager.GetUserInfo(token.AccessToken)
	if err != nil {
		logger.LOG.Error("获取百度用户信息失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取用户信息失败", nil))
		return
	}
	
	// 获取配额信息
	quota, _ := h.oauthManager.GetQuota(token.AccessToken)
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"user_info": userInfo,
		"quota":     quota,
	}))
}

// generateRandomState 生成随机state
func generateRandomState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// saveToken 保存Token到文件
func (h *CloudOAuthHandler) saveToken(userID string, token *BaiduToken) {
	tokenFile := filepath.Join(h.tokenDir, userID+".json")
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		logger.LOG.Error("序列化Token失败", "error", err)
		return
	}
	
	if err := os.WriteFile(tokenFile, data, 0644); err != nil {
		logger.LOG.Error("保存Token失败", "error", err)
	}
}

// loadTokens 加载保存的Token
func (h *CloudOAuthHandler) loadTokens() {
	files, err := os.ReadDir(h.tokenDir)
	if err != nil {
		return
	}
	
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		userID := file.Name()[:len(file.Name())-5]
		tokenFile := filepath.Join(h.tokenDir, file.Name())
		
		data, err := os.ReadFile(tokenFile)
		if err != nil {
			continue
		}
		
		var token BaiduToken
		if err := json.Unmarshal(data, &token); err != nil {
			continue
		}
		
		h.tokens[userID] = &token
	}
}

// GetBaiduToken 获取用户的百度Token（供其他服务调用）
func (h *CloudOAuthHandler) GetBaiduToken(userID string) (*BaiduToken, error) {
	h.mu.RLock()
	token, exists := h.tokens[userID]
	h.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("用户未登录百度网盘")
	}
	
	// 检查Token是否过期
	if time.Now().After(token.ExpiresAt) {
		// 尝试刷新Token
		newToken, err := h.oauthManager.RefreshToken(token.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("Token已过期，请重新登录")
		}
		
		// 更新Token
		token.AccessToken = newToken.AccessToken
		token.RefreshToken = newToken.RefreshToken
		token.ExpiresAt = newToken.ExpiresAt
		token.UpdatedAt = time.Now()
		
		h.mu.Lock()
		h.tokens[userID] = token
		h.mu.Unlock()
		
		h.saveToken(userID, token)
	}
	
	return token, nil
}
