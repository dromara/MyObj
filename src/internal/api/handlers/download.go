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

type DownloadHandler struct {
	service *service.DownloadService
	cache   cache.Cache
}

func NewDownloadHandler(service *service.DownloadService, cacheLocal cache.Cache) *DownloadHandler {
	return &DownloadHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (h *DownloadHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(h.cache,
		h.service.GetRepository().ApiKey(),
		h.service.GetRepository().User(),
		h.service.GetRepository().GroupPower(),
		h.service.GetRepository().Power())

	downloadGroup := c.Group("/download")
	{
		downloadGroup.Use(verify.Verify())
		// 创建离线下载任务
		downloadGroup.POST("/offline/create", middleware.PowerVerify("file:offLine"), h.CreateOfflineDownload)
		// 获取下载任务列表
		downloadGroup.GET("/list", middleware.PowerVerify("file:offLine"), h.GetTaskList)
		// 暂停下载任务
		downloadGroup.POST("/pause", middleware.PowerVerify("file:offLine"), h.PauseTask)
		// 恢复下载任务
		downloadGroup.POST("/resume", middleware.PowerVerify("file:offLine"), h.ResumeTask)
		// 取消下载任务
		downloadGroup.POST("/cancel", middleware.PowerVerify("file:offLine"), h.CancelTask)
		// 删除下载任务
		downloadGroup.POST("/delete", middleware.PowerVerify("file:offLine"), h.DeleteTask)
	}

	logger.LOG.Info("[路由] 下载路由注册完成✔️")
}

// CreateOfflineDownload 创建离线下载任务
func (h *DownloadHandler) CreateOfflineDownload(c *gin.Context) {
	req := new(request.CreateOfflineDownloadRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.CreateOfflineDownload(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "创建任务失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// GetTaskList 获取下载任务列表
func (h *DownloadHandler) GetTaskList(c *gin.Context) {
	req := new(request.DownloadTaskListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	// 默认查询所有状态
	if req.State == 0 && c.Query("state") == "" {
		req.State = -1
	}

	userID := c.GetString("userID")
	result, err := h.service.GetTaskList(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// PauseTask 暂停下载任务
func (h *DownloadHandler) PauseTask(c *gin.Context) {
	req := new(request.TaskOperationRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.PauseTask(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "暂停失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// ResumeTask 恢复下载任务
func (h *DownloadHandler) ResumeTask(c *gin.Context) {
	req := new(request.TaskOperationRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.ResumeTask(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "恢复失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// CancelTask 取消下载任务
func (h *DownloadHandler) CancelTask(c *gin.Context) {
	req := new(request.TaskOperationRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.CancelTask(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "取消失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// DeleteTask 删除下载任务
func (h *DownloadHandler) DeleteTask(c *gin.Context) {
	req := new(request.DeleteTaskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.DeleteTask(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "删除失败", err.Error()))
		return
	}

	c.JSON(200, result)
}
