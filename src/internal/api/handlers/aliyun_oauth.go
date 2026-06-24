package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"myobj/src/core/service"
	"myobj/src/pkg/auth"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/cloud/aliyun"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type AliyunOAuthState struct {
	State     string
	UserID    string
	CreatedAt time.Time
}

type AliyunOAuthHandler struct {
	oauthManager  *aliyun.AliyunOAuthManager
	tokenStore    *aliyun.AliyunTokenStore
	cache         cache.Cache
	accountService *service.CloudAccountService
	states        map[string]*AliyunOAuthState
	mu            sync.RWMutex
}

func NewAliyunOAuthHandler(cache cache.Cache, accountService *service.CloudAccountService) *AliyunOAuthHandler {
	return &AliyunOAuthHandler{
		oauthManager:  aliyun.NewAliyunOAuthManager("", "", ""),
		tokenStore:    aliyun.NewAliyunTokenStore(),
		cache:         cache,
		accountService: accountService,
		states:        make(map[string]*AliyunOAuthState),
	}
}

func (h *AliyunOAuthHandler) Router(c *gin.RouterGroup) {
	aliyun := c.Group("/cloud/aliyun")
	{
		aliyun.GET("/auth", h.GetAuthURL)
		aliyun.GET("/callback", h.HandleCallback)
		aliyun.GET("/status", h.GetStatus)
		aliyun.POST("/logout", h.Logout)
	}
}

func (h *AliyunOAuthHandler) GetAuthURL(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "请先登录MyObj系统", nil))
		return
	}

	state := genState()
	h.mu.Lock()
	h.states[state] = &AliyunOAuthState{State: state, UserID: userID, CreatedAt: time.Now()}
	h.mu.Unlock()

	authURL := h.oauthManager.GetAuthorizeURL(state)
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"auth_url": authURL, "state": state}))
}

func (h *AliyunOAuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(400, models.NewJsonResponse(400, "缺少参数", nil))
		return
	}

	h.mu.RLock()
	s, ok := h.states[state]
	h.mu.RUnlock()
	if !ok {
		c.JSON(400, models.NewJsonResponse(400, "无效state", nil))
		return
	}
	h.mu.Lock()
	delete(h.states, state)
	h.mu.Unlock()

	token, err := h.oauthManager.ExchangeCode(code)
	if err != nil {
		logger.LOG.Error("获取阿里云盘Token失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取Token失败: "+err.Error(), nil))
		return
	}

	h.tokenStore.Save(s.UserID, token)

	// 保存到数据库
	if h.accountService != nil {
		expiresIn := int(time.Until(token.ExpiresAt).Seconds())
		if err := h.accountService.SaveOAuthToken(nil, s.UserID, "aliyun", token.UserName, token.AccessToken, token.RefreshToken, expiresIn); err != nil {
			logger.LOG.Error("保存阿里云盘Token到数据库失败", "error", err)
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>阿里云盘授权成功</title>
<style>body{font-family:Arial,sans-serif;text-align:center;padding:50px}.ok{color:#4CAF50;font-size:24px}</style></head>
<body><div class="ok">✅ 阿里云盘授权成功！</div><p>用户: %s</p><p>有效期: %s</p>
<button onclick="window.close()">关闭</button>
<script>if(window.opener)window.opener.postMessage({type:'aliyun_auth_success'},'*');setTimeout(()=>window.close(),3000)</script>
</body></html>`, token.UserName, token.ExpiresAt.Format("2006-01-02 15:04:05"))
}

func (h *AliyunOAuthHandler) GetStatus(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"logged_in": false}))
		return
	}

	token, err := h.tokenStore.Get(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"logged_in": false}))
		return
	}

	if time.Now().After(token.ExpiresAt) {
		newToken, err := h.oauthManager.RefreshToken(token.RefreshToken)
		if err != nil {
			c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"logged_in": false, "error": "token_expired"}))
			return
		}
		token.AccessToken = newToken.AccessToken
		token.RefreshToken = newToken.RefreshToken
		token.ExpiresAt = newToken.ExpiresAt
		h.tokenStore.Save(userID, token)
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"logged_in":  true,
		"user_name":  token.UserName,
		"expires_at": token.ExpiresAt,
	}))
}

func (h *AliyunOAuthHandler) Logout(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "请先登录MyObj系统", nil))
		return
	}
	h.tokenStore.Delete(userID)
	// 从数据库删除
	if h.accountService != nil {
		h.accountService.DeleteAccount(nil, userID, "aliyun")
	}
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

func (h *AliyunOAuthHandler) GetTokenStore() *aliyun.AliyunTokenStore {
	return h.tokenStore
}

func (h *AliyunOAuthHandler) getUserID(c *gin.Context) string {
	// 从 Authorization header 解析
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return ""
	}
	// 从缓存获取实际JWT
	jwtToken, err := h.cache.Get(token)
	if err != nil {
		return ""
	}
	jwtStr, ok := jwtToken.(string)
	if !ok {
		return ""
	}
	// 解析JWT
	claims, err := auth.ParseToken(jwtStr)
	if err != nil {
		return ""
	}
	return claims.UserID
}

func genState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
