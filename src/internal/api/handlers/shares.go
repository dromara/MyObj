package handlers

import (
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
		share.GET("/download", s.GetShare)
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

// GetShare 获取分享
func (s *SharesHandler) GetShare(c *gin.Context) {
	token := c.Query("token")
	psw := c.Query("psw")

	// 如果没有提供密码，返回密码输入页面
	if psw == "" {
		c.HTML(200, "share_password.html", gin.H{"token": token})
		return
	}

	share := s.service.GetShare(token, psw)
	if share.Err != "" {
		c.HTML(200, "404.html", gin.H{"message": share.Err})
		return
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			logger.LOG.Error("删除临时文件失败", "error", err)
		}
	}(share.Temp)
	c.Header("Content-Disposition", "attachment; filename="+share.FileName)
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
