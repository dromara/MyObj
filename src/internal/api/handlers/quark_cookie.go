package handlers

import (
	"myobj/src/core/service"
	"myobj/src/pkg/cloud/quark"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// QuarkCookieHandler 夸克Cookie认证处理器
type QuarkCookieHandler struct {
	manager        *quark.CookieAuthManager
	accountService *service.CloudAccountService
}

// NewQuarkCookieHandler 创建夸克Cookie认证处理器
func NewQuarkCookieHandler(accountService *service.CloudAccountService) *QuarkCookieHandler {
	return &QuarkCookieHandler{
		manager:        quark.NewCookieAuthManager(),
		accountService: accountService,
	}
}

// Router 注册路由
func (h *QuarkCookieHandler) Router(c *gin.RouterGroup) {
	quark := c.Group("/cloud/quark/cookie")
	{
		// 保存Cookie
		quark.POST("/save", h.SaveCookie)
		// 获取Cookie状态
		quark.GET("/status", h.GetStatus)
		// 删除Cookie
		quark.POST("/delete", h.DeleteCookie)
		// 验证Cookie
		quark.POST("/validate", h.ValidateCookie)
		// 获取用户信息
		quark.GET("/userinfo", h.GetUserInfo)
	}
}

// SaveCookieRequest 保存Cookie请求
type SaveCookieRequest struct {
	Cookie string `json:"cookie" binding:"required"`
}

// SaveCookie 保存Cookie
func (h *QuarkCookieHandler) SaveCookie(c *gin.Context) {
	req := new(SaveCookieRequest)
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
		if err := h.accountService.SaveCookie(nil, userID, "quark", userName, req.Cookie); err != nil {
			// 记录日志但不阻断流程
			_ = err
		}
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"user_name": userName,
	}))
}

// GetStatus 获取Cookie状态
func (h *QuarkCookieHandler) GetStatus(c *gin.Context) {
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
func (h *QuarkCookieHandler) DeleteCookie(c *gin.Context) {
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
		h.accountService.DeleteAccount(nil, userID, "quark")
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// ValidateCookie 验证Cookie
func (h *QuarkCookieHandler) ValidateCookie(c *gin.Context) {
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

// GetUserInfo 获取用户信息
func (h *QuarkCookieHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	config, err := h.manager.GetConfig(userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "未配置夸克网盘Cookie", nil))
		return
	}
	
	// 验证Cookie是否仍然有效
	valid, userName, _ := h.manager.ValidateCookie(config.Cookie)
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"valid":      valid,
		"user_name":  userName,
		"updated_at": config.UpdatedAt,
	}))
}
