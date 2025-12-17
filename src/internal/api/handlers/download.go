package handlers

import (
	"context"
	"io"
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/download"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"
	"strconv"

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
		// 创建网盘文件下载任务
		downloadGroup.POST("/local/create", middleware.PowerVerify("file:download"), h.CreateLocalFileDownload)
		// 下载网盘文件
		downloadGroup.GET("/local/file/:taskID", middleware.PowerVerify("file:download"), h.DownloadLocalFile)
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

// CreateLocalFileDownload 创建网盘文件下载任务
// @Summary 创建网盘文件下载任务
// @Description 创建网盘文件下载任务，支持加密文件和分片文件
// @Tags 下载管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateLocalFileDownloadRequest true "下载请求"
// @Success 200 {object} models.JsonResponse{data=map[string]interface{}} "任务创建成功"
// @Failure 400 {object} models.JsonResponse "参数错误"
// @Failure 500 {object} models.JsonResponse "创建失败"
// @Router /download/local/create [post]
func (h *DownloadHandler) CreateLocalFileDownload(c *gin.Context) {
	req := new(request.CreateLocalFileDownloadRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := h.service.CreateLocalFileDownload(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "创建失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// DownloadLocalFile 下载网盘文件
// @Summary 下载网盘文件
// @Description 下载已准备完成的网盘文件，支持HTTP Range断点续传
// @Tags 下载管理
// @Produce octet-stream
// @Security BearerAuth
// @Param taskID path string true "任务ID"
// @Param Range header string false "Range请求头（例：bytes=0-1023）"
// @Success 200 {file} binary "文件流"
// @Success 206 {file} binary "部分文件流（Range请求）"
// @Failure 400 {object} models.JsonResponse "任务不存在或未准备完成"
// @Failure 500 {object} models.JsonResponse "下载失败"
// @Router /download/local/file/{taskID} [get]
func (h *DownloadHandler) DownloadLocalFile(c *gin.Context) {
	taskID := c.Param("taskID")
	userID := c.GetString("userID")

	// 1. 查询任务
	task, err := h.service.GetRepository().DownloadTask().GetByID(c.Request.Context(), taskID)
	if err != nil || task == nil {
		c.JSON(200, models.NewJsonResponse(400, "任务不存在", nil))
		return
	}

	// 2. 验证权限
	if task.UserID != userID {
		c.JSON(200, models.NewJsonResponse(403, "无权限访问", nil))
		return
	}

	// 3. 检查任务状态（必须是"已完成"状态）
	if task.State != 3 { // 3=已完成（网盘文件准备完成即可下载）
		c.JSON(200, models.NewJsonResponse(400, "任务未准备完成，请稍后再试", nil))
		return
	}

	// 4. 获取临时文件路径
	tempFilePath := task.Path // 使用Path字段存储的临时文件路径
	if tempFilePath == "" {
		c.JSON(200, models.NewJsonResponse(500, "文件不存在", nil))
		return
	}

	// 5. 打开文件
	file, err := os.Open(tempFilePath)
	if err != nil {
		logger.LOG.Error("打开文件失败", "error", err, "path", tempFilePath)
		c.JSON(200, models.NewJsonResponse(500, "文件不存在", nil))
		return
	}
	defer file.Close()

	// 6. 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.LOG.Error("获取文件信息失败", "error", err)
		c.JSON(200, models.NewJsonResponse(500, "获取文件信息失败", nil))
		return
	}

	fileSize := fileInfo.Size()

	// 7. 解析Range请求
	rangeHeader := c.GetHeader("Range")
	rangeInfo, err := download.ParseRangeHeader(rangeHeader, fileSize)
	if err != nil {
		logger.LOG.Warn("Range请求解析失败", "error", err, "range", rangeHeader)
		c.JSON(416, models.NewJsonResponse(416, "Range请求无效", nil))
		return
	}

	// 8. 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=\""+task.FileName+"\"")
	c.Header("Accept-Ranges", "bytes")

	if rangeInfo.IsRanged {
		// Range请求
		contentLength := rangeInfo.End - rangeInfo.Start + 1
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		c.Header("Content-Range", "bytes "+strconv.FormatInt(rangeInfo.Start, 10)+"-"+strconv.FormatInt(rangeInfo.End, 10)+"/"+strconv.FormatInt(fileSize, 10))
		c.Status(206)

		// 定位到起始位置
		if _, err := file.Seek(rangeInfo.Start, io.SeekStart); err != nil {
			logger.LOG.Error("文件定位失败", "error", err)
			c.JSON(500, models.NewJsonResponse(500, "文件读取失败", nil))
			return
		}

		// 读取指定范围的数据
		_, err = io.CopyN(c.Writer, file, contentLength)
		if err != nil && err != io.EOF {
			logger.LOG.Error("传输文件失败", "error", err)
			return
		}

		// Range请求不清理文件，等待完整请求或超时清理
		logger.LOG.Info("Range请求完成", "taskID", taskID, "range", rangeHeader)
	} else {
		// 完整文件请求
		c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
		c.Status(200)

		// 传输整个文件
		_, err = io.Copy(c.Writer, file)
		if err != nil && err != io.EOF {
			logger.LOG.Error("传输文件失败", "error", err)
			return
		}

		logger.LOG.Info("完整文件传输完成", "taskID", taskID, "fileName", task.FileName, "fileSize", fileSize)

		// 使用goroutine异步清理临时文件，避免阻塞响应
		go func() {
			// 更新任务状态为完成
			task.State = 3 // 3=完成
			task.FinishTime = custom_type.Now()
			h.service.GetRepository().DownloadTask().Update(context.Background(), task)

			// 如果是临时文件，则清理
			if download.IsTempPath(tempFilePath) {
				logger.LOG.Info("清理临时文件", "path", tempFilePath)
				os.RemoveAll(tempFilePath)
			} else {
				logger.LOG.Info("保留data文件，不清理", "path", tempFilePath)
			}
		}()
	}
	logger.LOG.Info("文件下载处理完成", "taskID", taskID, "fileName", task.FileName, "isRange", rangeInfo.IsRanged)
}
