package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"myobj/src/core/service"
	"myobj/src/pkg/cloud/pikpak"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type PikPakOAuthState struct {
	State     string
	UserID    string
	CreatedAt time.Time
}

type PikPakOAuthHandler struct {
	oauthManager  *pikpak.OAuthManager
	tokenStore    *pikpak.TokenStore
	accountService *service.CloudAccountService
	states        map[string]*PikPakOAuthState
	mu            sync.RWMutex
}

func NewPikPakOAuthHandler(accountService *service.CloudAccountService) *PikPakOAuthHandler {
	return &PikPakOAuthHandler{
		oauthManager:  pikpak.NewOAuthManager("", "", ""),
		tokenStore:    pikpak.NewTokenStore(),
		accountService: accountService,
		states:        make(map[string]*PikPakOAuthState),
	}
}

func (h *PikPakOAuthHandler) Router(c *gin.RouterGroup) {
	g := c.Group("/cloud/pikpak")
	{
		g.GET("/auth", h.GetAuthURL)
		g.GET("/callback", h.HandleCallback)
		g.GET("/status", h.GetStatus)
		g.POST("/logout", h.Logout)
	}
}

func (h *PikPakOAuthHandler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	state := genPikPakState()
	h.mu.Lock()
	h.states[state] = &PikPakOAuthState{State: state, UserID: userID, CreatedAt: time.Now()}
	h.mu.Unlock()
	authURL := h.oauthManager.GetAuthorizeURL(state)
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"auth_url": authURL, "state": state}))
}

func (h *PikPakOAuthHandler) HandleCallback(c *gin.Context) {
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
		logger.LOG.Error("获取PikPak Token失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取Token失败: "+err.Error(), nil))
		return
	}
	h.tokenStore.Save(s.UserID, token)

	// 保存到数据库
	if h.accountService != nil {
		expiresIn := int(time.Until(token.ExpiresAt).Seconds())
		if err := h.accountService.SaveOAuthToken(nil, s.UserID, "pikpak", token.UserName, token.AccessToken, token.RefreshToken, expiresIn); err != nil {
			logger.LOG.Error("保存PikPak Token到数据库失败", "error", err)
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>PikPak授权成功</title>
<style>body{font-family:Arial,sans-serif;text-align:center;padding:50px}.ok{color:#4CAF50;font-size:24px}</style></head>
<body><div class="ok">✅ PikPak授权成功！</div><p>有效期: %s</p>
<button onclick="window.close()">关闭</button>
<script>if(window.opener)window.opener.postMessage({type:'pikpak_auth_success'},'*');setTimeout(()=>window.close(),3000)</script>
</body></html>`, token.ExpiresAt.Format("2006-01-02 15:04:05"))
}

func (h *PikPakOAuthHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("userID")
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

func (h *PikPakOAuthHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	h.tokenStore.Delete(userID)
	// 从数据库删除
	if h.accountService != nil {
		h.accountService.DeleteAccount(nil, userID, "pikpak")
	}
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

func genPikPakState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
