package handlers

import (
	"strconv"

	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// CloudTransferHandler 云盘转存处理器
type CloudTransferHandler struct {
	service *service.CloudTransferService
	cache   cache.Cache
}

// NewCloudTransferHandler 创建云盘转存处理器
func NewCloudTransferHandler(service *service.CloudTransferService, cacheLocal cache.Cache) *CloudTransferHandler {
	return &CloudTransferHandler{
		service: service,
		cache:   cacheLocal,
	}
}

// Router 注册路由
func (h *CloudTransferHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddlewareFromFactory(h.cache, h.service.GetRepository())
	cloud := c.Group("/cloud")
	cloud.Use(verify.Verify())
	{
		// 转存单个文件
		cloud.POST("/transfer", h.TransferFile)
		// 获取转存状态
		cloud.GET("/transfer/status/:id", h.GetTransferStatus)
	}
}

// TransferFile 转存文件
// 支持单文件转存和任务全部文件转存
func (h *CloudTransferHandler) TransferFile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}

	var req struct {
		Provider string   `json:"provider" binding:"required"`
		ShareID  string   `json:"share_id" binding:"required"`
		ShareURL string   `json:"share_url"`
		SharePwd string   `json:"share_pwd"`
		FileIDs  []string `json:"file_ids"` // 为空则转存全部文件
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}

	result, err := h.service.TransferFromShare(c.Request.Context(), req.Provider, req.ShareID, req.ShareURL, req.SharePwd, req.FileIDs, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}

// GetTransferStatus 获取转存状态
func (h *CloudTransferHandler) GetTransferStatus(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, models.NewJsonResponse(401, "未授权", nil))
		return
	}

	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "invalid task id", nil))
		return
	}

	result, err := h.service.GetTransferStatus(c.Request.Context(), taskID, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", result))
}
