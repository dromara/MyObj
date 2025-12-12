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
	fileGroup := c.Group("/file")
	{
		fileGroup.Use(verify.Verify())
		// 获取文件列表
		fileGroup.GET("/list", middleware.PowerVerify("file:preview"), f.GetFileList)
		// 获取缩略图
		fileGroup.GET("/thumbnail/:fileId", middleware.PowerVerify("file:preview"), f.GetThumbnail)
		// 搜索当前用户文件
		fileGroup.GET("/search/user", middleware.PowerVerify("file:preview"), f.SearchUserFiles)
		fileGroup.GET("/file/search/public", middleware.PowerVerify("file:preview"), f.SearchPublicFiles)
		// 创建目录
		fileGroup.POST("/makeDir", middleware.PowerVerify("dir:create"), f.MakeDir)
		// 移动文件
		fileGroup.POST("/move", middleware.PowerVerify("file:move"), f.MoveFile)
		// 删除文件
		fileGroup.POST("/delete", middleware.PowerVerify("file:delete"), f.DeleteFile)
		// 获取虚拟路径
		fileGroup.GET("/virtualPath", middleware.PowerVerify("file:preview"), f.GetVirtualPath)
	}

	logger.LOG.Info("[路由] 文件路由注册完成✔️")
}

// Precheck 预检查
func (f *FileHandler) Precheck(c *gin.Context) {
	req := new(request.UploadPrecheckRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	req.UserID = c.GetString("userID")
	precheck, err := f.service.Precheck(req, f.cache)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, "预检查失败", err.Error()))
		return
	}
	c.JSON(200, precheck)
}

// SearchUserFiles 搜索当前用户文件
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

// GetFileList 获取文件列表
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

// GetThumbnail 获取文件缩略图
func (f *FileHandler) GetThumbnail(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(200, models.NewJsonResponse(400, "文件ID不能为空", nil))
		return
	}

	// 查询文件信息
	fileInfo, err := f.service.GetRepository().FileInfo().GetByID(c.Request.Context(), fileID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(404, "文件不存在", err.Error()))
		return
	}

	// 检查是否有缩略图
	if fileInfo.ThumbnailImg == "" {
		c.JSON(404, models.NewJsonResponse(404, "缩略图不存在", nil))
		return
	}

	// 设置响应头
	ext := filepath.Ext(fileInfo.ThumbnailImg)
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
	c.File(fileInfo.ThumbnailImg)

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

// DeleteFile 删除文件
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
