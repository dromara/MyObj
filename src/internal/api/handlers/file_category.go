package handlers

import (
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type FileCategoryHandler struct {
	service *service.FileCategoryService
	cache   cache.Cache
}

func NewFileCategoryHandler(s *service.FileCategoryService, cacheLocal cache.Cache) *FileCategoryHandler {
	return &FileCategoryHandler{
		service: s,
		cache:   cacheLocal,
	}
}

func (h *FileCategoryHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddlewareFromFactory(h.cache, h.service.GetRepository())

	catGroup := c.Group("/file/categories")
	{
		catGroup.Use(verify.Verify())
		catGroup.GET("", h.GetCategories)
		catGroup.GET("/stats", h.GetCategoryStats)
	}

	logger.LOG.Info("[路由] 文件分类路由注册完成✔️")
}

// GetCategories godoc
// @Summary 获取文件分类列表
// @Description 获取所有可用的文件分类及其扩展名
// @Tags 文件分类
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse "获取成功"
// @Router /file/categories [get]
func (h *FileCategoryHandler) GetCategories(c *gin.Context) {
	categories := h.service.GetAllCategoriesWithInfo()
	c.JSON(200, models.NewJsonResponse(200, "获取成功", categories))
}

// GetCategoryStats godoc
// @Summary 获取用户文件分类统计
// @Description 获取当前用户的文件分类统计信息
// @Tags 文件分类
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse{data=response.CategoryStatsResponse} "获取成功"
// @Router /file/categories/stats [get]
func (h *FileCategoryHandler) GetCategoryStats(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(200, models.NewJsonResponse(401, "未登录", nil))
		return
	}

	stats, err := h.service.GetUserCategoryStats(c.Request.Context(), userID)
	if err != nil {
		logger.LOG.Error("获取分类统计失败", "error", err, "userID", userID)
		c.JSON(200, models.NewJsonResponse(500, "获取分类统计失败", nil))
		return
	}

	c.JSON(200, models.NewJsonResponse(200, "获取成功", stats))
}
