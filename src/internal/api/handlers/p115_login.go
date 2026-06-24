package handlers

import (
	"myobj/src/pkg/cloud/p115"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type P115LoginHandler struct {
	manager *p115.LoginManager
}

func NewP115LoginHandler() *P115LoginHandler {
	return &P115LoginHandler{manager: p115.NewLoginManager()}
}

func (h *P115LoginHandler) Router(c *gin.RouterGroup) {
	g := c.Group("/cloud/115")
	{
		g.POST("/cookie", h.SaveCookie)
		g.GET("/status", h.GetStatus)
		g.POST("/logout", h.Logout)
	}
}

func (h *P115LoginHandler) SaveCookie(c *gin.Context) {
	var req struct {
		Cookie string `json:"cookie" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	valid, userName, err := h.manager.ValidateCookie(req.Cookie)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "验证失败: "+err.Error(), nil))
		return
	}
	if !valid {
		c.JSON(400, models.NewJsonResponse(400, "Cookie无效", nil))
		return
	}
	if err := h.manager.SaveCookie(userID, req.Cookie, userName); err != nil {
		c.JSON(500, models.NewJsonResponse(500, "保存失败: "+err.Error(), nil))
		return
	}
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"user_name": userName}))
}

func (h *P115LoginHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"logged_in": false}))
		return
	}
	cookie, err := h.manager.GetCookie(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{"logged_in": false}))
		return
	}
	valid, userName, _ := h.manager.ValidateCookie(cookie)
	c.JSON(200, models.NewJsonResponse(200, "ok", gin.H{
		"logged_in": true,
		"valid":     valid,
		"user_name": userName,
	}))
}

func (h *P115LoginHandler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	h.manager.DeleteCookie(userID)
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}
