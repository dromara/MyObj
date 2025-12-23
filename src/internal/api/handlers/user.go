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
		// API Key 相关路由
		r.POST("/apiKey/generate", middleware.PowerVerify("user:update"), u.GenerateApiKey)
		r.GET("/apiKey/list", middleware.PowerVerify("user:update"), u.ListApiKeys)
		r.POST("/apiKey/delete", middleware.PowerVerify("user:update"), u.DeleteApiKey)
	}
	logger.LOG.Info("[路由] 用户路由注册完成✔️")
}

// Login godoc
// @Summary 用户登录
// @Description 用户通过用户名和密码登录系统，返回JWT Token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UserLoginRequest true "登录请求"
// @Success 200 {object} models.JsonResponse{data=string} "token"
// @Failure 400 {object} models.JsonResponse "参数错误或登录失败"
// @Router /user/login [post]
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

// Register godoc
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body request.UserRegisterRequest true "注册请求"
// @Success 200 {object} models.JsonResponse "注册成功"
// @Failure 400 {object} models.JsonResponse "参数错误或注册失败"
// @Router /user/register [post]
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

// Challenge godoc
// @Summary 获取登录挑战值
// @Description 获取RSA公钥和挑战值，用于密码加密传输
// @Tags 用户管理
// @Produce json
// @Success 200 {object} models.JsonResponse{data=object} "挑战值信息"
// @Failure 400 {object} models.JsonResponse "获取失败"
// @Router /user/challenge [get]
func (u *UserHandler) Challenge(c *gin.Context) {
	challenge, err := u.service.Challenge()
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, challenge)
}

// SysInit godoc
// @Summary 获取系统初始化信息
// @Description 获取系统配置和初始化状态
// @Tags 系统管理
// @Produce json
// @Success 200 {object} models.JsonResponse "系统信息"
// @Failure 400 {object} models.JsonResponse "获取失败"
// @Router /user/sysInfo [get]
func (u *UserHandler) SysInit(c *gin.Context) {
	init, err := u.service.SysInit()
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, init)
}

// UpdateUser godoc
// @Summary 更新用户信息
// @Description 更新当前用户的基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UserUpdateRequest true "更新请求"
// @Success 200 {object} models.JsonResponse "更新成功"
// @Failure 400 {object} models.JsonResponse "参数错误或更新失败"
// @Router /user/updateUser [post]
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

// UpdatePassword godoc
// @Summary 修改登录密码
// @Description 修改用户登录密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UserUpdatePasswordRequest true "修改密码请求"
// @Success 200 {object} models.JsonResponse "修改成功"
// @Failure 400 {object} models.JsonResponse "参数错误或修改失败"
// @Router /user/updatePassword [post]
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

// SetFilePassword godoc
// @Summary 设置文件加密密码
// @Description 设置用户文件加密密码（首次设置）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UserUpdatePasswordRequest true "设置文件密码请求"
// @Success 200 {object} models.JsonResponse "设置成功"
// @Failure 400 {object} models.JsonResponse "参数错误或设置失败"
// @Router /user/setFilePassword [post]
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

// UserUpdateFilePassword godoc
// @Summary 修改文件加密密码
// @Description 修改用户文件加密密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UserUpdatePasswordRequest true "修改文件密码请求"
// @Success 200 {object} models.JsonResponse "修改成功"
// @Failure 400 {object} models.JsonResponse "参数错误或修改失败"
// @Router /user/updateFilePassword [post]
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

// GenerateApiKey godoc
// @Summary 生成API Key
// @Description 为用户生成新的API Key，用于API调用认证
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.GenerateApiKeyRequest true "生成API Key请求"
// @Success 200 {object} models.JsonResponse{data=object} "生成成功，返回API Key和公钥"
// @Failure 400 {object} models.JsonResponse "参数错误或生成失败"
// @Router /user/apiKey/generate [post]
func (u *UserHandler) GenerateApiKey(c *gin.Context) {
	req := new(request.GenerateApiKeyRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := u.service.GenerateApiKey(req, c.GetString("userID"))
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, result)
}

// ListApiKeys godoc
// @Summary 获取API Key列表
// @Description 获取当前用户的所有API Key列表（Key已掩码）
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.JsonResponse{data=[]object} "获取成功"
// @Failure 400 {object} models.JsonResponse "获取失败"
// @Router /user/apiKey/list [get]
func (u *UserHandler) ListApiKeys(c *gin.Context) {
	result, err := u.service.ListApiKeys(c.GetString("userID"))
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, result)
}

// DeleteApiKey godoc
// @Summary 删除API Key
// @Description 删除指定的API Key
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteApiKeyRequest true "删除API Key请求"
// @Success 200 {object} models.JsonResponse "删除成功"
// @Failure 400 {object} models.JsonResponse "参数错误或删除失败"
// @Failure 403 {object} models.JsonResponse "无权操作"
// @Failure 404 {object} models.JsonResponse "API Key不存在"
// @Router /user/apiKey/delete [post]
func (u *UserHandler) DeleteApiKey(c *gin.Context) {
	req := new(request.DeleteApiKeyRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	result, err := u.service.DeleteApiKey(req, c.GetString("userID"))
	if err != nil {
		c.JSON(400, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, result)
}
