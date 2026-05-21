package handlers

import (
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/core/service"
	"myobj/src/pkg/audit"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EnterpriseSpaceHandler struct {
	service *service.EnterpriseSpaceService
	cache   cache.Cache
}

func NewEnterpriseSpaceHandler(service *service.EnterpriseSpaceService, cacheLocal cache.Cache) *EnterpriseSpaceHandler {
	return &EnterpriseSpaceHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (h *EnterpriseSpaceHandler) CreateDir(c *gin.Context) {
	req := new(request.CreateSharedDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.CreateDir(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "创建目录失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_mkdir",
			TargetType:   "enterprise_shared_path",
			TargetName:   req.Name,
			TargetPath:   fmt.Sprintf("%d", req.ParentID),
			Detail:       fmt.Sprintf("创建目录「%s」", req.Name),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) ListFiles(c *gin.Context) {
	req := new(request.SharedFileListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.ListFiles(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "获取文件列表失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) UploadPrecheck(c *gin.Context) {
	req := new(request.SharedUploadPrecheckRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.UploadPrecheck(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "预检失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) UploadFile(c *gin.Context) {
	req := new(request.SharedFileUploadRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, "获取上传文件失败", err.Error()))
		return
	}
	defer file.Close()

	userID := c.GetString("userID")
	result, err := h.service.UploadFile(req, file, header, userID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "上传失败", err.Error()))
		return
	}

	if result.Code == 200 {
		// 审计日志
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       userID,
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_upload",
			TargetType:   "enterprise_shared_file",
			TargetName:   header.Filename,
			TargetPath:   fmt.Sprintf("%d", req.PathID),
			Detail:       fmt.Sprintf("上传文件「%s」到企业共享空间", header.Filename),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}

	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) DeleteFile(c *gin.Context) {
	req := new(request.DeleteSharedFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.DeleteFile(req.ID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "删除文件失败: "+err.Error(), nil))
		}
		return
	}

	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_delete",
			TargetType:   "enterprise_shared_file",
			TargetName:   req.ID,
			Detail:       fmt.Sprintf("删除企业共享空间文件 %s", req.ID),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}

	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) DownloadFile(c *gin.Context) {
	fileID := c.Query("id")
	if fileID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}

	result, err := h.service.DownloadFile(fileID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "下载文件失败: "+err.Error(), nil))
		}
		return
	}

	// 如果查询成功，直接提供文件下载
	if result.Code == 200 {
		data, ok := result.Data.(map[string]interface{})
		if ok {
			filePath, _ := data["file_path"].(string)
			fileName, _ := data["file_name"].(string)
			mime, _ := data["mime"].(string)
			if filePath != "" {
				audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
					ID:           uuid.New().String(),
					UserID:       c.GetString("userID"),
					UserName:     getEnterpriseUserName(c),
					EnterpriseID: c.GetString("enterpriseID"),
					Action:       "enterprise_space_download",
					TargetType:   "enterprise_shared_file",
					TargetName:   fileName,
					TargetPath:   filePath,
					Detail:       fmt.Sprintf("下载企业共享空间文件 %s", fileName),
					IP:           c.ClientIP(),
					CreatedAt:    custom_type.Now(),
				})
				encodedName := url.PathEscape(fileName)
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encodedName))
				c.Header("Content-Type", mime)
				c.File(filePath)
				return
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

func (h *EnterpriseSpaceHandler) GetSpaceUsage(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.GetSpaceUsage(enterpriseID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "获取空间用量失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) DeleteDir(c *gin.Context) {
	req := new(request.DeleteSharedDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.DeleteDir(req.ID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "删除目录失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_delete_dir",
			TargetType:   "enterprise_shared_path",
			TargetName:   fmt.Sprintf("%d", req.ID),
			TargetPath:   fmt.Sprintf("%d", req.ID),
			Detail:       fmt.Sprintf("删除企业共享空间目录 %d", req.ID),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) RenameFile(c *gin.Context) {
	req := new(request.RenameSharedFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.RenameFile(req.ID, req.Name, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "重命名文件失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_rename",
			TargetType:   "enterprise_shared_file",
			TargetName:   req.Name,
			Detail:       fmt.Sprintf("重命名文件为「%s」", req.Name),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

func (h *EnterpriseSpaceHandler) RenameDir(c *gin.Context) {
	req := new(request.RenameSharedDirRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.RenameDir(req.ID, req.Name, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "重命名目录失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_rename_dir",
			TargetType:   "enterprise_shared_path",
			TargetName:   req.Name,
			Detail:       fmt.Sprintf("重命名目录为「%s」", req.Name),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

// ========== 新增功能 Handler ==========

// PreviewFile 文件预览（inline 返回）
func (h *EnterpriseSpaceHandler) PreviewFile(c *gin.Context) {
	fileID := c.Query("id")
	if fileID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.DownloadFile(fileID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "预览失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		data, ok := result.Data.(map[string]interface{})
		if ok {
			// 加密文件不允许预览
			if isEnc, _ := data["is_enc"].(bool); isEnc {
				c.JSON(200, models.NewJsonResponse(400, "加密文件不支持预览，请下载后查看", nil))
				return
			}
			filePath, _ := data["file_path"].(string)
			fileName, _ := data["file_name"].(string)
			mime, _ := data["mime"].(string)
			if filePath != "" {
				audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
					ID:           uuid.New().String(),
					UserID:       c.GetString("userID"),
					UserName:     getEnterpriseUserName(c),
					EnterpriseID: c.GetString("enterpriseID"),
					Action:       "enterprise_space_preview",
					TargetType:   "enterprise_shared_file",
					TargetName:   fileName,
					TargetPath:   filePath,
					Detail:       fmt.Sprintf("预览企业共享空间文件 %s", fileName),
					IP:           c.ClientIP(),
					CreatedAt:    custom_type.Now(),
				})
				encodedName := url.PathEscape(fileName)
				c.Header("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", encodedName))
				c.Header("Content-Type", mime)
				c.File(filePath)
				return
			}
		}
	}
	c.JSON(200, result)
}

// GetThumbnail 获取文件缩略图
func (h *EnterpriseSpaceHandler) GetThumbnail(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	thumbnailPath, err := h.service.GetThumbnailPath(fileID, c.GetString("userID"))
	if err != nil {
		c.JSON(200, models.NewJsonResponse(404, "缩略图不存在", nil))
		return
	}
	c.File(thumbnailPath)
}

// SearchFiles 搜索企业空间文件
func (h *EnterpriseSpaceHandler) SearchFiles(c *gin.Context) {
	req := new(request.SearchEnterpriseFilesRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.SearchFiles(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "搜索失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

// GetPathTree 获取企业空间目录树
func (h *EnterpriseSpaceHandler) GetPathTree(c *gin.Context) {
	enterpriseID := c.Query("enterprise_id")
	if enterpriseID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.GetPathTree(enterpriseID, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "获取目录树失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

// MoveFile 移动企业空间文件
func (h *EnterpriseSpaceHandler) MoveFile(c *gin.Context) {
	req := new(request.MoveEnterpriseFileRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.MoveFile(req.FileID, req.TargetPath, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "移动文件失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_move",
			TargetType:   "enterprise_shared_file",
			TargetName:   req.FileID,
			TargetPath:   fmt.Sprintf("%d", req.TargetPath),
			Detail:       fmt.Sprintf("移动文件 %s 到目录 %d", req.FileID, req.TargetPath),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

// PackageCreate 创建打包下载任务
func (h *EnterpriseSpaceHandler) PackageCreate(c *gin.Context) {
	req := new(request.EnterprisePackageCreateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.PackageCreate(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "创建打包任务失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_package",
			TargetType:   "enterprise_shared_file",
			TargetName:   fmt.Sprintf("%d files", len(req.FileIDs)),
			Detail:       fmt.Sprintf("打包下载 %d 个文件", len(req.FileIDs)),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

// PackageProgress 查询打包进度
func (h *EnterpriseSpaceHandler) PackageProgress(c *gin.Context) {
	packageID := c.Query("package_id")
	if packageID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.PackageProgress(packageID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询进度失败", nil))
		return
	}
	c.JSON(200, result)
}

// PackageDownload 下载打包文件
func (h *EnterpriseSpaceHandler) PackageDownload(c *gin.Context) {
	packageID := c.Query("package_id")
	if packageID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	filePath, fileName, err := h.service.GetPackageFile(packageID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(404, "打包文件不存在或已过期", nil))
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/zip")
	c.File(filePath)
}

// ExtractCheck 检测解压冲突
func (h *EnterpriseSpaceHandler) ExtractCheck(c *gin.Context) {
	req := new(request.EnterpriseExtractCheckRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.ExtractCheck(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "检测失败: "+err.Error(), nil))
		}
		return
	}
	c.JSON(200, result)
}

// ExtractStart 开始解压
func (h *EnterpriseSpaceHandler) ExtractStart(c *gin.Context) {
	req := new(request.ExtractStartRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.ExtractStart(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "解压失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_extract",
			TargetType:   "enterprise_shared_file",
			TargetName:   req.FileID,
			Detail:       fmt.Sprintf("解压文件 %s", req.FileID),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

// ExtractProgress 查询解压进度
func (h *EnterpriseSpaceHandler) ExtractProgress(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := h.service.ExtractProgress(taskID)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(500, "查询进度失败", nil))
		return
	}
	c.JSON(200, result)
}

// CreateShare 创建企业文件分享链接
func (h *EnterpriseSpaceHandler) CreateShare(c *gin.Context) {
	req := new(request.CreateEnterpriseShareRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(200, models.NewJsonResponse(400, "参数错误", err.Error()))
		return
	}
	result, err := h.service.CreateShare(req, c.GetString("userID"))
	if err != nil {
		if result != nil {
			c.JSON(200, result)
		} else {
			c.JSON(200, models.NewJsonResponse(500, "创建分享失败: "+err.Error(), nil))
		}
		return
	}
	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:           uuid.New().String(),
			UserID:       c.GetString("userID"),
			UserName:     getEnterpriseUserName(c),
			EnterpriseID: c.GetString("enterpriseID"),
			Action:       "enterprise_space_share",
			TargetType:   "enterprise_shared_file",
			TargetName:   req.FileID,
			Detail:       fmt.Sprintf("分享文件 %s", req.FileID),
			IP:           c.ClientIP(),
			CreatedAt:    custom_type.Now(),
		})
	}
	c.JSON(200, result)
}

func getEnterpriseUserName(c *gin.Context) string {
	if userLogin, exists := c.Get("userLogin"); exists {
		if info, ok := userLogin.(response.UserLoginResponse); ok && info.User != nil {
			if info.User.Name != "" {
				return info.User.Name
			}
			return info.User.UserName
		}
	}
	return ""
}
