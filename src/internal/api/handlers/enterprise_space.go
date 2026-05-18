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
		c.JSON(200, result)
		return
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
		c.JSON(200, result)
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
		c.JSON(200, result)
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

	// 审计日志
	audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		UserName:   getEnterpriseUserName(c),
		Action:     "enterprise_space_upload",
		TargetType: "enterprise_shared_file",
		TargetName: header.Filename,
		Detail:     fmt.Sprintf("上传文件「%s」到企业共享空间", header.Filename),
		IP:         c.ClientIP(),
		CreatedAt:  custom_type.Now(),
	})

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
		c.JSON(200, result)
		return
	}

	if result.Code == 200 {
		audit.Record(h.service.GetRepository().DB(), &models.AuditLog{
			ID:         uuid.New().String(),
			UserID:     c.GetString("userID"),
			UserName:   getEnterpriseUserName(c),
			Action:     "enterprise_space_delete",
			TargetType: "enterprise_shared_file",
			TargetName: req.ID,
			Detail:     fmt.Sprintf("删除企业共享空间文件 %s", req.ID),
			IP:         c.ClientIP(),
			CreatedAt:  custom_type.Now(),
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
		c.JSON(200, result)
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
					ID:         uuid.New().String(),
					UserID:     c.GetString("userID"),
					UserName:   getEnterpriseUserName(c),
					Action:     "enterprise_space_download",
					TargetType: "enterprise_shared_file",
					TargetName: fileName,
					Detail:     fmt.Sprintf("下载企业共享空间文件 %s", fileName),
					IP:         c.ClientIP(),
					CreatedAt:  custom_type.Now(),
				})
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
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
		c.JSON(200, result)
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
		c.JSON(200, result)
		return
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
		c.JSON(200, result)
		return
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
		c.JSON(200, result)
		return
	}
	c.JSON(200, result)
}

func getEnterpriseUserName(c *gin.Context) string {
	if userLogin, exists := c.Get("userLogin"); exists {
		if info, ok := userLogin.(response.UserLoginResponse); ok && info.User != nil {
			return info.User.Name
		}
	}
	return ""
}
