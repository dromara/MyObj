package handlers

import (
	"fmt"
	"net/url"
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"

	"github.com/gin-gonic/gin"
)

type SharesHandler struct {
	service *service.SharesService
	cache   cache.Cache
}

func NewSharesHandler(service *service.SharesService, cacheLocal cache.Cache) *SharesHandler {
	return &SharesHandler{
		service: service,
		cache:   cacheLocal,
	}
}
func (s *SharesHandler) GetRepository() *impl.RepositoryFactory {
	return s.service.GetRepository()
}

func (s *SharesHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(s.cache,
		s.service.GetRepository().ApiKey(),
		s.service.GetRepository().User(),
		s.service.GetRepository().GroupPower(),
		s.service.GetRepository().Power())
	share := c.Group("/share")
	{
		share.GET("/info", s.GetShareInfo)      // 获取分享信息（不触发下载）
		share.GET("/download", s.DownloadShare) // 下载分享文件（GET请求，直接触发下载）
	}
	ver := c.Group("/share")
	ver.Use(verify.Verify())
	{
		// 创建分享
		ver.POST("/create", middleware.PowerVerify("file:share"), s.CreateShare)
		// 获取分享列表
		ver.GET("/list", middleware.PowerVerify("file:share"), s.GetShareList)
		// 删除分享
		ver.POST("/delete", middleware.PowerVerify("file:share"), s.DeleteShare)
		// 修改分享密码
		ver.POST("/updatePassword", middleware.PowerVerify("file:share"), s.UpdateSharePassword)
	}
}

// CreateShare 创建分享
func (s *SharesHandler) CreateShare(c *gin.Context) {
	req := new(request.CreateShareRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	createShare, err := s.service.CreateShare(req, c.GetString("userID"))
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, createShare)
}

// GetShareInfo 获取分享信息（不触发下载）
func (s *SharesHandler) GetShareInfo(c *gin.Context) {
	token := c.Query("token")
	password := c.Query("password") // 可选，如果有密码则必需
	
	if token == "" {
		c.JSON(400, models.NewJsonResponse(400, "token参数不能为空", nil))
		return
	}

	shareInfo, err := s.service.GetShareInfo(token, password)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}

	c.JSON(200, models.NewJsonResponse(200, "ok", shareInfo))
}

// DownloadShare 下载分享文件（GET请求，直接触发浏览器下载）
func (s *SharesHandler) DownloadShare(c *gin.Context) {
	token := c.Query("token")
	password := c.Query("password") // 可选，如果有密码则必需

	if token == "" {
		c.JSON(400, models.NewJsonResponse(400, "token参数不能为空", nil))
		return
	}

	// 调用服务下载分享文件
	share := s.service.DownloadShare(token, password)
	if share.Err != "" {
		c.JSON(400, models.NewJsonResponse(400, share.Err, nil))
		return
	}
	
	// 检查返回的数据是否有效
	if share.Path == "" {
		logger.LOG.Error("分享文件路径为空", "token", token)
		c.JSON(400, models.NewJsonResponse(400, "文件路径无效", nil))
		return
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(share.Path); os.IsNotExist(err) {
		logger.LOG.Error("分享文件不存在", "path", share.Path, "error", err)
		c.JSON(404, models.NewJsonResponse(404, "文件不存在或已被删除", nil))
		return
	}
	
	defer func(name string) {
		if name != "" {
			err := os.RemoveAll(name)
			if err != nil {
				logger.LOG.Error("删除临时文件失败", "error", err)
			}
		}
	}(share.Temp)
	
	// 处理文件名，确保特殊字符被正确编码
	fileName := share.FileName
	if fileName == "" {
		fileName = "download"
	}
	// 使用 RFC 5987 格式编码文件名，支持中文和特殊字符
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, 
		fileName, url.QueryEscape(fileName)))
	c.Header("Content-Type", "application/octet-stream")
	
	// 使用 c.File 发送文件，直接触发浏览器下载
	c.File(share.Path)
}

// GetShareList 获取分享列表
func (s *SharesHandler) GetShareList(c *gin.Context) {
	userID := c.GetString("userID")
	shareList, err := s.service.GetShareList(userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, shareList)
}

// DeleteShare 删除分享
func (s *SharesHandler) DeleteShare(c *gin.Context) {
	req := new(request.DeleteShareRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	userID := c.GetString("userID")
	deleteShare, err := s.service.DeleteShare(req.ID, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, deleteShare)
}

// UpdateSharePassword 修改分享密码
func (s *SharesHandler) UpdateSharePassword(c *gin.Context) {
	req := new(request.UpdateSharePasswordRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	userID := c.GetString("userID")
	updatePassword, err := s.service.UpdateSharePassword(req.ID, req.Password, userID)
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, updatePassword)
}
