package handlers

import (
	"errors"
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	service *service.UserService
	cache   cache.Cache
}

func NewUserHandler(service *service.UserService, cacheLocal cache.Cache) *UserHandler {
	return &UserHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (u *UserHandler) Router(c *gin.RouterGroup) {
	c.POST("/user/login", u.Login)
	c.POST("/user/register", u.Register)
	c.GET("/user/sysInfo", u.SysInit)
	c.GET("/user/challenge", u.Challenge)

	verify := middleware.NewAuthMiddleware(u.cache,
		u.service.GetRepository().ApiKey(),
		u.service.GetRepository().User(),
		u.service.GetRepository().GroupPower(),
		u.service.GetRepository().Power())

	r := c.Group("/user")
	r.Use(verify.Verify())
	{
		r.POST("/updateUser", middleware.PowerVerify("user:update"), u.UpdateUser)
		r.POST("/updateUserElse", middleware.PowerVerify("user:update:else"), u.UpdateUser)
		r.POST("/updatePassword", middleware.PowerVerify("user:update:password"), u.UpdatePassword)
		r.POST("/setFilePassword", middleware.PowerVerify("file:update:filePassword"), u.SetFilePassword)
		r.POST("/updateFilePassword", middleware.PowerVerify("file:update:filePassword"), u.UserUpdateFilePassword)
	}
	logger.LOG.Info("[路由] 用户路由注册完成✔️")
}

func (u *UserHandler) Login(c *gin.Context) {
	req := new(request.UserLoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	login, err := u.service.Login(req.Username, req.Password, req.Challenge)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户不存在", nil))
			return
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, login)
}

func (u *UserHandler) Register(c *gin.Context) {
	req := new(request.UserRegisterRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	register, err := u.service.Register(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户已存在", nil))
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
	}
	c.JSON(200, register)
}

// Challenge 密码传递挑战
func (u *UserHandler) Challenge(c *gin.Context) {
	challenge, err := u.service.Challenge()
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, challenge)
}

func (u *UserHandler) SysInit(c *gin.Context) {
	init, err := u.service.SysInit()
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, init)
}

func (u *UserHandler) UpdateUser(c *gin.Context) {
	req := new(request.UserUpdateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	req.ID = c.GetString("userID")
	update, err := u.service.UpdateUser(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户不存在", nil))
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, update)
}

func (u *UserHandler) UpdatePassword(c *gin.Context) {
	req := new(request.UserUpdatePasswordRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	req.ID = c.GetString("userID")
	update, err := u.service.UpdatePassword(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户不存在", nil))
			return
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
	}
	c.JSON(200, update)
}

func (u *UserHandler) SetFilePassword(c *gin.Context) {
	req := new(request.UserUpdatePasswordRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	req.ID = c.GetString("userID")
	update, err := u.service.UpdatePassword(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户不存在", nil))
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
	}
	c.JSON(200, update)
}

func (u *UserHandler) UserUpdateFilePassword(c *gin.Context) {
	req := new(request.UserUpdatePasswordRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	req.ID = c.GetString("userID")
	update, err := u.service.UpdateFilePassword(req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, models.NewJsonResponse(400, "用户不存在", nil))
		}
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
	}
	c.JSON(200, update)
}
