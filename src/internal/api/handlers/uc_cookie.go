package handlers

import (
	"myobj/src/core/service"
	"myobj/src/pkg/cloud/uc"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// UCCookieHandler UC网盘Cookie认证处理器
type UCCookieHandler struct {
	manager        *uc.CookieAuthManager
	accountService *service.CloudAccountService
}

// NewUCCookieHandler 创建UC网盘Cookie认证处理器
func NewUCCookieHandler(accountService *service.CloudAccountService) *UCCookieHandler {
	return &UCCookieHandler{
		manager:        uc.NewCookieAuthManager(),
		accountService: accountService,
	}
}

// Router 注册路由
func (h *UCCookieHandler) Router(c *gin.RouterGroup) {
	uc := c.Group("/cloud/uc/cookie")
	{
		// 保存Cookie
		uc.POST("/save", h.SaveCookie)
		// 获取Cookie状态
		uc.GET("/status", h.GetStatus)
		// 删除Cookie
		uc.POST("/delete", h.DeleteCookie)
		// 验证Cookie
		uc.POST("/validate", h.ValidateCookie)
	}
}

// UCSaveCookieRequest 保存Cookie请求
type UCSaveCookieRequest struct {
	Cookie string `json:"cookie" binding:"required"`
}

// SaveCookie 保存Cookie
func (h *UCCookieHandler) SaveCookie(c *gin.Context) {
	req := new(UCSaveCookieRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	// 验证Cookie
	valid, userName, err := h.manager.ValidateCookie(req.Cookie)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "验证Cookie失败: "+err.Error(), nil))
		return
	}
	
	if !valid {
		c.JSON(400, models.NewJsonResponse(400, "Cookie无效", nil))
		return
	}
	
	// 保存Cookie
	if err := h.manager.SaveCookie(userID, req.Cookie, userName); err != nil {
		c.JSON(500, models.NewJsonResponse(500, "保存Cookie失败: "+err.Error(), nil))
		return
	}

	// 保存到数据库
	if h.accountService != nil {
		if err := h.accountService.SaveCookie(nil, userID, "uc", userName, req.Cookie); err != nil {
			_ = err
		}
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"user_name": userName,
	}))
}

// GetStatus 获取Cookie状态
func (h *UCCookieHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"configured": false,
		}))
		return
	}
	
	config, err := h.manager.GetConfig(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"configured": false,
		}))
		return
	}
	
	// 验证Cookie是否仍然有效
	valid, _, _ := h.manager.ValidateCookie(config.Cookie)
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"configured": true,
		"valid":      valid,
		"user_name":  config.UserName,
		"updated_at": config.UpdatedAt,
	}))
}

// DeleteCookie 删除Cookie
func (h *UCCookieHandler) DeleteCookie(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	if err := h.manager.DeleteCookie(userID); err != nil {
		c.JSON(500, models.NewJsonResponse(500, "删除Cookie失败: "+err.Error(), nil))
		return
	}

	// 从数据库删除
	if h.accountService != nil {
		h.accountService.DeleteAccount(nil, userID, "uc")
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// ValidateCookie 验证Cookie
func (h *UCCookieHandler) ValidateCookie(c *gin.Context) {
	req := new(SaveCookieRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	valid, userName, err := h.manager.ValidateCookie(req.Cookie)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "验证Cookie失败: "+err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"valid":     valid,
		"user_name": userName,
	}))
}
