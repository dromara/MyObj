package handlers

import (
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	service *service.FileService
	cache   cache.Cache
}

func NewFileHandler(service *service.FileService, cacheLocal cache.Cache) *FileHandler {
	return &FileHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (f *FileHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(f.cache,
		f.service.GetRepository().ApiKey(),
		f.service.GetRepository().User(),
		f.service.GetRepository().GroupPower(),
		f.service.GetRepository().Power())

	// 公开路由（不需要验证）
	publicGroup := c.Group("/file")
	{
		// 公开文件列表
		publicGroup.GET("/public/list", f.PublicFileList)
	}

	// 需要验证的路由
	fileGroup := c.Group("/file")
	{
		fileGroup.Use(verify.Verify())
		// 预检接口
		fileGroup.POST("/upload/precheck", middleware.PowerVerify("file:upload"), f.Precheck)
		// 上传进度查询接口
		fileGroup.GET("/upload/progress", middleware.PowerVerify("file:upload"), f.GetUploadProgress)
		// 查询未完成的上传任务列表
		fileGroup.GET("/upload/uncompleted", middleware.PowerVerify("file:upload"), f.ListUncompletedUploads)
		// 查询过期的上传任务列表
		fileGroup.GET("/upload/expired", middleware.PowerVerify("file:upload"), f.ListExpiredUploads)
		// 删除上传任务
		fileGroup.POST("/upload/delete", middleware.PowerVerify("file:upload"), f.DeleteUploadTask)
		// 延期过期任务（恢复任务）
		fileGroup.POST("/upload/renew", middleware.PowerVerify("file:upload"), f.RenewExpiredTask)
		// 清理过期的上传任务（用户可清理自己的，系统自动清理所有）
		fileGroup.POST("/upload/clean-expired", middleware.PowerVerify("file:upload"), f.CleanExpiredUploads)
		// 文件上传接口
		fileGroup.POST("/upload", middleware.PowerVerify("file:upload"), f.UploadFile)
		// 获取文件列表
		fileGroup.GET("/list", middleware.PowerVerify("file:preview"), f.GetFileList)
		// 获取缩略图
		fileGroup.GET("/thumbnail/:fileId", middleware.PowerVerify("file:preview"), f.GetThumbnail)
		// 搜索当前用户文件
		fileGroup.GET("/search/user", middleware.PowerVerify("file:preview"), f.SearchUserFiles)
		// 搜索公开文件
		fileGroup.GET("/search/public", middleware.PowerVerify("file:preview"), f.SearchPublicFiles)
		// 创建目录
		fileGroup.POST("/makeDir", middleware.PowerVerify("dir:create"), f.MakeDir)
		// 移动文件
		fileGroup.POST("/move", middleware.PowerVerify("file:move"), f.MoveFile)
		// 删除文件
		fileGroup.POST("/delete", middleware.PowerVerify("file:delete"), f.DeleteFile)
		// 重命名文件（业务逻辑已验证文件所有权，无需额外权限验证）
		fileGroup.POST("/rename", f.RenameFile)
		// 重命名目录（业务逻辑已验证目录所有权，无需额外权限验证）
		fileGroup.POST("/renameDir", f.RenameDir)
		// 删除目录（业务逻辑已验证目录所有权，无需额外权限验证）
		fileGroup.POST("/deleteDir", f.DeleteDir)
		// 设置文件公开状态（业务逻辑已验证文件所有权和加密状态，无需额外权限验证）
		fileGroup.POST("/setPublic", f.SetFilePublic)
		// 获取虚拟路径
		fileGroup.GET("/virtualPath", middleware.PowerVerify("file:preview"), f.GetVirtualPath)
	}

	logger.LOG.Info("[路由] 文件路由注册完成✔️")
}

// Precheck godoc
// @Summary 文件上传预检
// @Description 上传前的预检查，检查空间、秒传可能性，返回预检ID
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UploadPrecheckRequest true "预检请求"
// @Success 200 {object} models.JsonResponse{data=string} "预检ID"
// @Success 200 {object} models.JsonResponse{message=string} "秒传成功"
// @Failure 400 {object} models.JsonResponse "预检失败"
// @Router /file/upload/precheck [post]
func (f *FileHandler) Precheck(c *gin.Context) {
	req := new(request.UploadPrecheckRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	req.UserID = c.GetString("userID")
	precheck, err := f.service.Precheck(req, f.cache)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, "预检查失败", err.Error()))
		return
	}
	c.JSON(200, precheck)
}

// SearchUserFiles godoc
// @Summary 搜索当前用户文件
// @Description 根据关键词搜索当前用户的文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} models.JsonResponse{data=object} "搜索结果"
// @Failure 500 {object} models.JsonResponse "搜索失败"
// @Router /file/search/user [get]
func (f *FileHandler) SearchUserFiles(c *gin.Context) {
	req := new(request.FileSearchRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	userID := c.GetString("userID")
	result, err := f.service.SearchUserFiles(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "搜索失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// SearchPublicFiles 搜索公开文件（广场）
func (f *FileHandler) SearchPublicFiles(c *gin.Context) {
	req := new(request.FileSearchRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.SearchPublicFiles(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "搜索失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// GetFileList godoc
// @Summary 获取文件列表
// @Description 获取当前用户指定目录下的文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param virtualPath query string false "虚拟路径"
// @Param page query int true "页码" minimum(1)
// @Param pageSize query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} models.JsonResponse{data=object} "文件列表"
// @Failure 500 {object} models.JsonResponse "获取失败"
// @Router /file/list [get]
func (f *FileHandler) GetFileList(c *gin.Context) {
	req := new(request.FileListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	userID := c.GetString("userID")
	result, err := f.service.GetFileList(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "获取失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// sendThumbnailResponse 发送缩略图响应（提取的公共逻辑）
func (f *FileHandler) sendThumbnailResponse(c *gin.Context, thumbnailPath string) {
	// 检查是否有缩略图
	if thumbnailPath == "" {
		c.JSON(404, models.NewJsonResponse(404, "缩略图不存在", nil))
		return
	}

	// 设置响应头
	ext := filepath.Ext(thumbnailPath)
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // 缓存1天
	c.File(thumbnailPath)
}

// GetThumbnail 获取文件缩略图
func (f *FileHandler) GetThumbnail(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(200, models.NewJsonResponse(400, "文件ID不能为空", nil))
		return
	}

	userID := c.GetString("userID")
	ctx := c.Request.Context()

	// 先通过 uf_id 查询 user_files 表，获取真实的 file_id
	// 因为前端传递的是 uf_id（用户文件关联表的ID），而不是 file_info 表的 id
	userFile, err := f.service.GetRepository().UserFiles().GetByUserIDAndUfID(ctx, userID, fileID)
	if err != nil {
		// 如果通过 uf_id 查询失败，尝试直接作为 file_id 查询（兼容旧版本）
		fileInfo, err2 := f.service.GetRepository().FileInfo().GetByID(ctx, fileID)
		if err2 != nil {
			c.JSON(200, models.NewJsonResponse(404, "文件不存在", err.Error()))
			return
		}
		// 发送缩略图响应
		f.sendThumbnailResponse(c, fileInfo.ThumbnailImg)
		return
	}

	// 通过 user_files 获取到的 file_id 查询 file_info
	fileInfo, err := f.service.GetRepository().FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(404, "文件不存在", err.Error()))
		return
	}

	// 发送缩略图响应
	f.sendThumbnailResponse(c, fileInfo.ThumbnailImg)
}

// MakeDir 创建目录
func (f *FileHandler) MakeDir(c *gin.Context) {
	req := new(request.MakeDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	userID := c.GetString("userID")
	makeDir, err := f.service.MakeDir(req, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "创建目录失败", err.Error()))
		return
	}
	c.JSON(200, makeDir)
}

// MoveFile 移动文件
func (f *FileHandler) MoveFile(c *gin.Context) {
	req := new(request.MoveFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	moveFile, err := f.service.MoveFile(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "移动文件失败", err.Error()))
		return
	}
	c.JSON(200, moveFile)
}

// GetVirtualPath 获取虚拟路径
func (f *FileHandler) GetVirtualPath(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := f.service.GetVirtualPath(userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "获取虚拟路径失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// DeleteFile godoc
// @Summary 删除文件
// @Description 将文件移动到回收站（软删除）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteFileRequest true "删除请求"
// @Success 200 {object} models.JsonResponse{data=object} "删除结果"
// @Failure 500 {object} models.JsonResponse "删除失败"
// @Router /file/delete [post]
func (f *FileHandler) DeleteFile(c *gin.Context) {
	req := new(request.DeleteFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.DeleteFiles(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "删除文件失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// RenameFile godoc
// @Summary 重命名文件
// @Description 重命名用户文件
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.RenameFileRequest true "重命名请求"
// @Success 200 {object} models.JsonResponse{data=object} "重命名成功"
// @Failure 400 {object} models.JsonResponse "参数错误或重命名失败"
// @Failure 404 {object} models.JsonResponse "文件不存在"
// @Router /file/rename [post]
func (f *FileHandler) RenameFile(c *gin.Context) {
	req := new(request.RenameFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.RenameFile(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "重命名文件失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// RenameDir godoc
// @Summary 重命名目录
// @Description 重命名用户目录，并自动更新子目录和文件的路径
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.RenameDirRequest true "重命名请求"
// @Success 200 {object} models.JsonResponse{data=object} "重命名成功"
// @Failure 400 {object} models.JsonResponse "参数错误或重命名失败"
// @Failure 404 {object} models.JsonResponse "目录不存在"
// @Failure 403 {object} models.JsonResponse "无权访问"
// @Router /file/renameDir [post]
func (f *FileHandler) RenameDir(c *gin.Context) {
	req := new(request.RenameDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.RenameDir(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "重命名目录失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// DeleteDir godoc
// @Summary 删除目录
// @Description 删除目录及其下的所有文件和子目录（文件会移动到回收站）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteDirRequest true "删除目录请求"
// @Success 200 {object} models.JsonResponse{data=object} "删除成功"
// @Failure 400 {object} models.JsonResponse "参数错误或根目录不能删除"
// @Failure 404 {object} models.JsonResponse "目录不存在"
// @Failure 403 {object} models.JsonResponse "无权访问"
// @Router /file/deleteDir [post]
func (f *FileHandler) DeleteDir(c *gin.Context) {
	req := new(request.DeleteDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.DeleteDir(req, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "删除目录失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// SetFilePublic godoc
// @Summary 设置文件公开状态
// @Description 设置文件是否公开（加密文件不能设置为公开）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.SetFilePublicRequest true "设置公开状态请求"
// @Success 200 {object} models.JsonResponse{data=object} "设置成功"
// @Failure 400 {object} models.JsonResponse "参数错误或加密文件不能公开"
// @Failure 404 {object} models.JsonResponse "文件不存在"
// @Router /file/setPublic [post]
func (f *FileHandler) SetFilePublic(c *gin.Context) {
	req := new(request.SetFilePublicRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		logger.LOG.Error("参数错误", "err", err)
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.SetFilePublic(req, c.GetString("userID"))
	if err != nil {
		logger.LOG.Error("设置文件公开状态失败", "err", err)
		c.JSON(200, models.NewJsonResponse(500, "设置文件公开状态失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// UploadFile godoc
// @Summary 文件上传
// @Description 支持小文件直传和大文件分片上传
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param precheck_id formData string true "预检ID"
// @Param file formData file true "文件数据"
// @Param chunk_index formData int false "分片索引"
// @Param total_chunks formData int false "总分片数"
// @Param chunk_md5 formData string false "分片MD5"
// @Param is_enc formData boolean false "是否加密"
// @Success 200 {object} models.JsonResponse{data=object} "上传成功"
// @Failure 400 {object} models.JsonResponse "上传失败"
// @Router /file/upload [post]
func (f *FileHandler) UploadFile(c *gin.Context) {
	// 1. 解析请求参数
	req := new(request.FileUploadRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	// 2. 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, "获取上传文件失败", err.Error()))
		return
	}
	defer file.Close()

	// 3. 调用 Service 处理上传
	userID := c.GetString("userID")
	result, err := f.service.UploadFile(req, file, header, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "上传失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// GetUploadProgress godoc
// @Summary 查询上传进度
// @Description 根据预检ID查询文件上传进度
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param precheck_id query string true "预检ID"
// @Success 200 {object} models.JsonResponse{data=response.UploadProgressResponse} "进度信息"
// @Failure 400 {object} models.JsonResponse "参数错误"
// @Failure 404 {object} models.JsonResponse "预检信息不存在"
// @Router /file/upload/progress [get]
func (f *FileHandler) GetUploadProgress(c *gin.Context) {
	req := new(request.UploadProgressRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := f.service.GetUploadProgress(req, userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}

	c.JSON(200, result)
}

// PublicFileList 广场公开文件列表
// @Summary 获取广场公开文件列表
// @Description 获取广场公开文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param request body request.PublicFileListRequest true "请求参数"
// @Success 200 {object} models.JsonResponse{data=object} "成功"
// @Failure 500 {object} models.JsonResponse "失败"
// @Router /file/public/list [get]
func (f *FileHandler) PublicFileList(c *gin.Context) {
	req := new(request.PublicFileListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := f.service.PublicFileList(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "获取文件列表失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// ListUncompletedUploads godoc
// @Summary 查询未完成的上传任务列表
// @Description 查询当前用户所有未完成的上传任务（用于断点续传）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse{data=[]object} "未完成的上传任务列表"
// @Failure 500 {object} models.JsonResponse "查询失败"
// @Router /file/upload/uncompleted [get]
func (f *FileHandler) ListUncompletedUploads(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := f.service.ListUncompletedUploads(userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// DeleteUploadTask godoc
// @Summary 删除上传任务
// @Description 删除指定的上传任务（从数据库中删除记录）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteUploadTaskRequest true "删除请求"
// @Success 200 {object} models.JsonResponse "删除成功"
// @Failure 400 {object} models.JsonResponse "参数错误"
// @Failure 500 {object} models.JsonResponse "删除失败"
// @Router /file/upload/delete [post]
func (f *FileHandler) DeleteUploadTask(c *gin.Context) {
	req := new(request.DeleteUploadTaskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := f.service.DeleteUploadTask(req.TaskID, userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "删除失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// CleanExpiredUploads godoc
// @Summary 清理过期的上传任务
// @Description 清理过期的未完成上传任务。如果提供 userID 参数，则只清理该用户的过期任务；如果不提供，则清理所有用户的过期任务（系统自动清理）
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "用户ID（可选，不提供则清理所有用户的过期任务）"
// @Success 200 {object} models.JsonResponse{data=object} "清理结果"
// @Failure 500 {object} models.JsonResponse "清理失败"
// @Router /file/upload/clean-expired [post]
func (f *FileHandler) CleanExpiredUploads(c *gin.Context) {
	// 获取当前用户ID（用户清理自己的任务）
	userID := c.GetString("userID")

	// 如果提供了 user_id 查询参数，使用该参数（用于系统自动清理）
	if queryUserID := c.Query("user_id"); queryUserID != "" {
		userID = queryUserID
	}

	result, err := f.service.CleanExpiredUploads(userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "清理失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// ListExpiredUploads godoc
// @Summary 查询过期的上传任务列表
// @Description 查询当前用户所有过期的上传任务
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse{data=[]object} "过期的上传任务列表"
// @Failure 500 {object} models.JsonResponse "查询失败"
// @Router /file/upload/expired [get]
func (f *FileHandler) ListExpiredUploads(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := f.service.ListExpiredUploads(userID)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "查询失败", err.Error()))
		return
	}
	c.JSON(200, result)
}

// RenewExpiredTask godoc
// @Summary 延期过期任务（恢复任务）
// @Description 延期过期的上传任务，延长过期时间使其可以继续上传
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.RenewExpiredTaskRequest true "延期请求"
// @Success 200 {object} models.JsonResponse{data=object} "延期成功"
// @Failure 400 {object} models.JsonResponse "参数错误"
// @Failure 403 {object} models.JsonResponse "无权操作"
// @Failure 404 {object} models.JsonResponse "任务不存在"
// @Failure 500 {object} models.JsonResponse "延期失败"
// @Router /file/upload/renew [post]
func (f *FileHandler) RenewExpiredTask(c *gin.Context) {
	req := new(request.RenewExpiredTaskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	userID := c.GetString("userID")
	result, err := f.service.RenewExpiredTask(req.TaskID, userID, req.Days)
	if err != nil {
		c.JSON(500, models.NewJsonResponse(500, "延期失败", err.Error()))
		return
	}
	c.JSON(200, result)
}
