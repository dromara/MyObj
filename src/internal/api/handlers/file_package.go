package handlers

import (
	"io"
	"myobj/src/core/domain/request"
	"myobj/src/pkg/download"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
// 使用流式传输避免浏览器拦截302跳转
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

	// 安全验证：规范化路径并检查是否在允许的目录内
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		logger.LOG.Error("路径规范化失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "文件路径无效", nil))
		return
	}

	// 检查路径是否包含危险字符（路径遍历攻击防护）
	cleanPath := filepath.Clean(absPath)
	if cleanPath != absPath || strings.Contains(absPath, "..") {
		logger.LOG.Warn("检测到可疑路径", "packageID", req.PackageID, "userID", userID)
		c.JSON(403, models.NewJsonResponse(403, "文件路径无效", nil))
		return
	}

	// 验证路径是否在系统临时目录内（防止路径遍历攻击）
	tempDir, err := filepath.Abs(os.TempDir())
	if err != nil {
		logger.LOG.Error("获取临时目录失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "系统错误", nil))
		return
	}

	// 检查文件路径是否在临时目录内
	if !strings.HasPrefix(cleanPath, tempDir) {
		logger.LOG.Warn("文件路径不在允许的目录内", "packageID", req.PackageID, "userID", userID)
		c.JSON(403, models.NewJsonResponse(403, "文件路径无效", nil))
		return
	}

	// 打开文件
	file, err := os.Open(cleanPath)
	if err != nil {
		logger.LOG.Error("打开打包文件失败", "error", err, "packageID", req.PackageID)
		c.JSON(404, models.NewJsonResponse(404, "文件不存在", nil))
		return
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			logger.LOG.Warn("关闭打包文件失败", "error", err, "packageID", req.PackageID)
		}
	}(file)

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		logger.LOG.Error("获取文件信息失败", "error", err)
		c.JSON(500, models.NewJsonResponse(500, "获取文件信息失败", nil))
		return
	}

	fileSize := fileInfo.Size()

	// 解析Range请求（支持断点续传）
	rangeHeader := c.GetHeader("Range")
	rangeInfo, err := download.ParseRangeHeader(rangeHeader, fileSize)
	if err != nil {
		logger.LOG.Warn("Range请求解析失败", "error", err, "range", rangeHeader)
		// 如果Range解析失败，继续传输完整文件
		rangeInfo = &download.FileRangeInfo{
			IsRanged:  false,
			Start:     0,
			End:       fileSize - 1,
			TotalSize: fileSize,
		}
	}

	// 设置响应头（避免浏览器拦截）
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="`+fileName+`"`)
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-Content-Type-Options", "nosniff") // 防止浏览器MIME类型嗅探
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	if rangeInfo.IsRanged {
		// Range请求（断点续传）
		contentLength := rangeInfo.End - rangeInfo.Start + 1
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		c.Header("Content-Range", "bytes "+strconv.FormatInt(rangeInfo.Start, 10)+"-"+strconv.FormatInt(rangeInfo.End, 10)+"/"+strconv.FormatInt(fileSize, 10))
		c.Status(206) // Partial Content

		// 定位到起始位置
		if _, err := file.Seek(rangeInfo.Start, io.SeekStart); err != nil {
			logger.LOG.Error("文件定位失败", "error", err)
			c.JSON(500, models.NewJsonResponse(500, "文件读取失败", nil))
			return
		}

		// 传输指定范围的数据
		_, err = io.CopyN(c.Writer, file, contentLength)
		if err != nil && err != io.EOF {
			logger.LOG.Error("传输文件失败", "error", err)
			return
		}
	} else {
		// 完整文件请求
		c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
		c.Status(200) // OK

		// 传输完整文件
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			logger.LOG.Error("传输文件失败", "error", err)
			return
		}
	}

	// 日志中不记录完整路径，只记录文件名和大小，避免泄露服务器结构
	logger.LOG.Info("打包文件下载完成", "packageID", req.PackageID, "fileName", fileName, "fileSize", fileSize)
}
