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

// GetRecycledList godoc
// @Summary 获取回收站列表
// @Description 获取当前用户回收站中的文件列表
// @Tags 回收站
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int true "页码" minimum(1)
// @Param pageSize query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} models.JsonResponse{data=object} "回收站列表"
// @Failure 500 {object} models.JsonResponse "获取失败"
// @Router /recycled/list [get]
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

// RestoreFile godoc
// @Summary 还原文件
// @Description 从回收站还原文件到原位置
// @Tags 回收站
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.RestoreFileRequest true "还原请求"
// @Success 200 {object} models.JsonResponse "还原成功"
// @Failure 500 {object} models.JsonResponse "还原失败"
// @Router /recycled/restore [post]
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

// DeletePermanently godoc
// @Summary 永久删除文件
// @Description 从回收站永久删除文件，不可恢复
// @Tags 回收站
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteRecycledRequest true "删除请求"
// @Success 200 {object} models.JsonResponse "删除成功"
// @Failure 500 {object} models.JsonResponse "删除失败"
// @Router /recycled/delete [post]
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

// EmptyRecycled godoc
// @Summary 清空回收站
// @Description 清空当前用户回收站中的所有文件
// @Tags 回收站
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse "清空成功"
// @Failure 500 {object} models.JsonResponse "清空失败"
// @Router /recycled/empty [post]
func (h *RecycledHandler) EmptyRecycled(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := h.service.EmptyRecycled(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "清空回收站失败", err.Error()))
		return
	}

	c.JSON(200, result)
}
