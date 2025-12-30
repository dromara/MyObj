package handlers

import (
	"myobj/src/core/domain/request"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// CreatePackage 创建打包下载任务
func (f *FileHandler) CreatePackage(c *gin.Context) {
	req := new(request.PackageCreateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}

	userID := c.GetString("userID")
	result, err := f.service.CreatePackage(req, userID)
	if err != nil {
		logger.LOG.Error("创建打包任务失败", "error", err)
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, result)
}

// GetPackageProgress 获取打包进度
func (f *FileHandler) GetPackageProgress(c *gin.Context) {
	req := new(request.PackageProgressRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}

	userID := c.GetString("userID")
	result, err := f.service.GetPackageProgress(req.PackageID, userID)
	if err != nil {
		logger.LOG.Error("获取打包进度失败", "error", err)
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, result)
}

// DownloadPackage 下载打包文件
func (f *FileHandler) DownloadPackage(c *gin.Context) {
	req := new(request.PackageDownloadRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}

	userID := c.GetString("userID")
	filePath, fileName, err := f.service.DownloadPackage(req.PackageID, userID)
	if err != nil {
		logger.LOG.Error("下载打包文件失败", "error", err)
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.File(filePath)
}

