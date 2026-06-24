package handlers

import (
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type CloudAccountHandler struct {
	service *service.CloudAccountService
	cache   cache.Cache
}

func NewCloudAccountHandler(s *service.CloudAccountService, cacheLocal cache.Cache) *CloudAccountHandler {
	return &CloudAccountHandler{service: s, cache: cacheLocal}
}

func (h *CloudAccountHandler) GetRepository() *impl.RepositoryFactory {
	return h.service.GetRepository()
}

func (h *CloudAccountHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddlewareFromFactory(h.cache, h.service.GetRepository())
	g := c.Group("/cloud/accounts")
	g.Use(verify.Verify())
	{
		g.GET("", h.ListAccounts)
		g.GET("/status", h.GetAllStatus)
		g.DELETE("/:provider", h.DeleteAccount)
	}
}

func (h *CloudAccountHandler) ListAccounts(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	accounts, err := h.service.ListAccounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "获取失败", nil))
		return
	}
	c.JSON(200, models.NewJsonResponse(200, "ok", accounts))
}

func (h *CloudAccountHandler) GetAllStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	status := h.service.CheckAllStatus(c.Request.Context(), userID)
	c.JSON(200, models.NewJsonResponse(200, "ok", status))
}

func (h *CloudAccountHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	provider := c.Param("provider")
	if err := h.service.DeleteAccount(c.Request.Context(), userID, provider); err != nil {
		c.JSON(500, models.NewJsonResponse(500, "删除失败", nil))
		return
	}
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}
