package handlers

import (
	"myobj/src/pkg/cloud/tianyi"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// TianyiLoginHandler 天翼云盘登录处理器
type TianyiLoginHandler struct {
	manager *tianyi.LoginManager
}

// NewTianyiLoginHandler 创建天翼云盘登录处理器
func NewTianyiLoginHandler() *TianyiLoginHandler {
	return &TianyiLoginHandler{
		manager: tianyi.NewLoginManager(),
	}
}

// Router 注册路由
func (h *TianyiLoginHandler) Router(c *gin.RouterGroup) {
	tianyi := c.Group("/cloud/tianyi")
	{
		// 用户名密码登录
		tianyi.POST("/login", h.Login)
		// 获取登录状态
		tianyi.GET("/status", h.GetStatus)
		// 登出
		tianyi.POST("/logout", h.Logout)
		// 获取用户空间信息
		tianyi.GET("/userinfo", h.GetUserInfo)
	}
}

// TianyiLoginRequest 登录请求
type TianyiLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (h *TianyiLoginHandler) Login(c *gin.Context) {
	req := new(TianyiLoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	// 登录
	userName, err := h.manager.Login(userID, req.Username, req.Password)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "登录失败: "+err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"user_name": userName,
	}))
}

// GetStatus 获取登录状态
func (h *TianyiLoginHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"logged_in": false,
		}))
		return
	}
	
	config, err := h.manager.GetConfig(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
			"logged_in": false,
		}))
		return
	}
	
	// 验证Cookie是否仍然有效
	valid, _ := h.manager.ValidateCookie(config.Cookie)
	
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"logged_in":  true,
		"valid":      valid,
		"user_name":  config.UserName,
		"updated_at": config.UpdatedAt,
	}))
}

// Logout 登出
func (h *TianyiLoginHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	if err := h.manager.Logout(userID); err != nil {
		c.JSON(500, models.NewJsonResponse(500, "登出失败: "+err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// GetUserInfo 获取用户空间信息
func (h *TianyiLoginHandler) GetUserInfo(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	userInfo, err := h.manager.GetUserSizeInfo(userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "获取用户信息失败: "+err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", userInfo))
}
