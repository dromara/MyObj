package handlers

import (
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type RecycledHandler struct {
	service *service.RecycledService
	cache   cache.Cache
}

func NewRecycledHandler(service *service.RecycledService, cacheLocal cache.Cache) *RecycledHandler {
	return &RecycledHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (h *RecycledHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(h.cache,
		h.service.GetRepository().ApiKey(),
		h.service.GetRepository().User(),
		h.service.GetRepository().GroupPower(),
		h.service.GetRepository().Power())

	recycled := c.Group("/recycled")
	recycled.Use(verify.Verify())
	{
		// 获取回收站列表
		recycled.GET("/list", middleware.PowerVerify("file:preview"), h.GetRecycledList)
		// 还原文件
		recycled.POST("/restore", middleware.PowerVerify("file:delete"), h.RestoreFile)
		// 永久删除文件
		recycled.POST("/delete", middleware.PowerVerify("file:delete"), h.DeletePermanently)
		// 清空回收站
		recycled.POST("/empty", middleware.PowerVerify("file:delete"), h.EmptyRecycled)
	}

	logger.LOG.Info("[路由] 回收站路由注册完成✔️")
}

// GetRecycledList 获取回收站列表
func (h *RecycledHandler) GetRecycledList(c *gin.Context) {
	req := new(request.RecycledListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.GetRecycledList(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "获取回收站列表失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// RestoreFile 还原文件
func (h *RecycledHandler) RestoreFile(c *gin.Context) {
	req := new(request.RestoreFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.RestoreFile(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "还原文件失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// DeletePermanently 永久删除文件
func (h *RecycledHandler) DeletePermanently(c *gin.Context) {
	req := new(request.DeleteRecycledRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.DeletePermanently(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "永久删除失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// EmptyRecycled 清空回收站
func (h *RecycledHandler) EmptyRecycled(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := h.service.EmptyRecycled(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "清空回收站失败", err.Error()))
		return
	}

	c.JSON(200, result)
}
