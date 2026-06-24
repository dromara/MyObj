package handlers

import (
	"strconv"

	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// CloudHandler 云盘处理器
type CloudHandler struct {
	service *service.CloudService
	cache   cache.Cache
}

// NewCloudHandler 创建云盘处理器
func NewCloudHandler(service *service.CloudService, cacheLocal cache.Cache) *CloudHandler {
	return &CloudHandler{
		service: service,
		cache:   cacheLocal,
	}
}

// GetRepository 获取仓储工厂
func (h *CloudHandler) GetRepository() *impl.RepositoryFactory {
	return nil // 云盘服务不直接使用Repository
}

// Router 注册路由
func (h *CloudHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddlewareFromFactory(h.cache, h.service.GetRepository())
	cloud := c.Group("/cloud")
	cloud.Use(verify.Verify())
	{
		// 获取支持的云盘列表
		cloud.GET("/providers", h.GetProviders)
		// 解析分享链接
		cloud.POST("/parse", h.ParseShareLink)
		// 列出分享文件
		cloud.POST("/list", h.ListShareFiles)
		// 获取分享文件树（支持递归）
		cloud.POST("/tree", h.GetShareFileTree)
		// 保存分享文件到本地
		cloud.POST("/save", h.SaveShareFiles)
		// 下载分享文件
		cloud.POST("/download", h.DownloadShareFile)
		// 获取任务状态
		cloud.GET("/task/:id", h.GetTaskStatus)
		// 获取任务列表
		cloud.GET("/tasks", h.GetUserTasks)
	}
}

// GetProviders 获取支持的云盘列表
func (h *CloudHandler) GetProviders(c *gin.Context) {
	providers := h.service.GetSupportedProviders()
	c.JSON(200, models.NewJsonResponse(200, "ok", providers))
}

// ParseShareLink 解析分享链接
func (h *CloudHandler) ParseShareLink(c *gin.Context) {
	req := new(request.ParseShareLinkRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	result, err := h.service.ParseShareLink(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// ListShareFiles 列出分享文件
func (h *CloudHandler) ListShareFiles(c *gin.Context) {
	req := new(request.ListShareFilesRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	result, err := h.service.ListShareFiles(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// DownloadShareFile 下载分享文件
func (h *CloudHandler) DownloadShareFile(c *gin.Context) {
	req := new(request.DownloadShareFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	err := h.service.DownloadShareFile(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", nil))
}

// SaveShareFiles 保存分享文件到本地
func (h *CloudHandler) SaveShareFiles(c *gin.Context) {
	req := new(request.SaveShareFilesRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	result, err := h.service.SaveShareFiles(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// GetShareFileTree 获取分享文件树
func (h *CloudHandler) GetShareFileTree(c *gin.Context) {
	req := new(request.GetShareFileTreeRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	result, err := h.service.GetShareFileTree(c.Request.Context(), req, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// GetTaskStatus 获取任务状态
func (h *CloudHandler) GetTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "invalid task id", nil))
		return
	}
	
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	result, err := h.service.GetTaskStatus(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// GetUserTasks 获取用户任务列表
func (h *CloudHandler) GetUserTasks(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	result, err := h.service.GetUserTasks(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	
	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}
