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

	"myobj/src/pkg/cloud/quark"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// QuarkOAuthState 临时存储OAuth状态
type QuarkOAuthState struct {
	State     string    `json:"state"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// QuarkToken 夸克Token存储
type QuarkToken struct {
	UserID       string    `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserName     string    `json:"user_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// QuarkOAuthHandler 夸克OAuth处理器
type QuarkOAuthHandler struct {
	oauthManager *quark.OAuthManager
	states       map[string]*QuarkOAuthState
	tokens       map[string]*QuarkToken
	mu           sync.RWMutex
	tokenDir     string
}

// NewQuarkOAuthHandler 创建夸克OAuth处理器
func NewQuarkOAuthHandler() *QuarkOAuthHandler {
	handler := &QuarkOAuthHandler{
		oauthManager: quark.NewOAuthManager("", "", ""),
		states:       make(map[string]*QuarkOAuthState),
		tokens:       make(map[string]*QuarkToken),
		tokenDir:     "./data/quark_tokens",
	}
	
	// 创建token存储目录
	os.MkdirAll(handler.tokenDir, 0755)
	
	// 加载已保存的token
	handler.loadTokens()
	
	return handler
}

// Router 注册OAuth路由
func (h *QuarkOAuthHandler) Router(c *gin.RouterGroup) {
	oauth := c.Group("/cloud/quark")
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

// GetAuthURL 获取夸克授权URL
func (h *QuarkOAuthHandler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	// 生成随机state
	state := generateQuarkRandomState()
	
	// 保存state
	h.mu.Lock()
	h.states[state] = &QuarkOAuthState{
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
func (h *QuarkOAuthHandler) HandleCallback(c *gin.Context) {
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
		logger.LOG.Error("获取夸克Token失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取Token失败: "+err.Error(), nil))
		return
	}
	
	// 获取用户信息
	userInfo, _ := h.oauthManager.GetUserInfo(token.AccessToken)
	userName := ""
	if name, ok := userInfo["name"].(string); ok {
		userName = name
	}
	
	// 保存Token
	quarkToken := &QuarkToken{
		UserID:       oauthState.UserID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
		UserName:     userName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	h.mu.Lock()
	h.tokens[oauthState.UserID] = quarkToken
	h.mu.Unlock()
	
	// 保存到文件
	h.saveToken(oauthState.UserID, quarkToken)
	
	// 返回成功页面
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>夸克网盘授权成功</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #4CAF50; font-size: 24px; margin-bottom: 20px; }
        .info { color: #666; margin-bottom: 10px; }
        .btn { background: #4CAF50; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; }
        .btn:hover { background: #45a049; }
    </style>
</head>
<body>
    <div class="success">✅ 夸克网盘授权成功！</div>
    <div class="info">用户: %s</div>
    <div class="info">Token有效期: %s</div>
    <br>
    <button class="btn" onclick="window.close()">关闭窗口</button>
    <script>
        // 通知父窗口授权成功
        if (window.opener) {
            window.opener.postMessage({type: 'quark_auth_success', user: '%s'}, '*');
        }
        // 3秒后自动关闭
        setTimeout(() => window.close(), 3000);
    </script>
</body>
</html>
`, userName, token.ExpiresAt.Format("2006-01-02 15:04:05"), userName)
}

// GetLoginStatus 获取登录状态
func (h *QuarkOAuthHandler) GetLoginStatus(c *gin.Context) {
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
			logger.LOG.Error("刷新夸克Token失败", "error", err)
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
func (h *QuarkOAuthHandler) Logout(c *gin.Context) {
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
	
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// GetUserInfo 获取用户信息
func (h *QuarkOAuthHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	h.mu.RLock()
	token, exists := h.tokens[userID]
	h.mu.RUnlock()
	
	if !exists {
		c.JSON(400, models.NewJsonResponse(400, "未登录夸克网盘", nil))
		return
	}
	
	// 获取用户信息
	userInfo, err := h.oauthManager.GetUserInfo(token.AccessToken)
	if err != nil {
		logger.LOG.Error("获取夸克用户信息失败", "error", err)
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

// generateQuarkRandomState 生成随机state
func generateQuarkRandomState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// saveToken 保存Token到文件
func (h *QuarkOAuthHandler) saveToken(userID string, token *QuarkToken) {
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
func (h *QuarkOAuthHandler) loadTokens() {
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
		
		var token QuarkToken
		if err := json.Unmarshal(data, &token); err != nil {
			continue
		}
		
		h.tokens[userID] = &token
	}
}

// GetQuarkToken 获取用户的夸克Token（供其他服务调用）
func (h *QuarkOAuthHandler) GetQuarkToken(userID string) (*QuarkToken, error) {
	h.mu.RLock()
	token, exists := h.tokens[userID]
	h.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("用户未登录夸克网盘")
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
