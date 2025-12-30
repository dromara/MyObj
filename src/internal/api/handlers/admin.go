package handlers

import (
	"myobj/src/core/domain/request"
	"myobj/src/core/service"
	"myobj/src/internal/api/middleware"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service *service.AdminService
	cache   cache.Cache
}

func NewAdminHandler(service *service.AdminService, cacheLocal cache.Cache) *AdminHandler {
	return &AdminHandler{
		service: service,
		cache:   cacheLocal,
	}
}

func (a *AdminHandler) Router(c *gin.RouterGroup) {
	verify := middleware.NewAuthMiddleware(a.cache,
		a.service.GetRepository().ApiKey(),
		a.service.GetRepository().User(),
		a.service.GetRepository().GroupPower(),
		a.service.GetRepository().Power())

	admin := c.Group("/admin")
	admin.Use(verify.Verify())
	admin.Use(middleware.AdminVerify()) // 管理员权限验证
	{
		// 用户管理
		admin.GET("/user/list", a.UserList)
		admin.POST("/user/create", a.CreateUser)
		admin.POST("/user/update", a.UpdateUser)
		admin.POST("/user/delete", a.DeleteUser)
		admin.POST("/user/toggle-state", a.ToggleUserState)

		// 组管理
		admin.GET("/group/list", a.GroupList)
		admin.POST("/group/create", a.CreateGroup)
		admin.POST("/group/update", a.UpdateGroup)
		admin.POST("/group/delete", a.DeleteGroup)

		// 权限管理
		admin.GET("/power/list", a.PowerList)
		admin.POST("/power/create", a.CreatePower)
		admin.POST("/power/update", a.UpdatePower)
		admin.POST("/power/delete", a.DeletePower)
		admin.POST("/power/batch-delete", a.BatchDeletePower)
		admin.POST("/power/assign", a.AssignPower)
		admin.GET("/power/group-powers", a.GetGroupPowers)

		// 磁盘管理
		admin.GET("/disk/list", a.DiskList)
		admin.POST("/disk/create", a.CreateDisk)
		admin.POST("/disk/update", a.UpdateDisk)
		admin.POST("/disk/delete", a.DeleteDisk)

		// 系统配置
		admin.GET("/system/config", a.GetSystemConfig)
		admin.POST("/system/update-config", a.UpdateSystemConfig)
	}

	logger.LOG.Info("[路由] 管理路由注册完成✔️")
}

// ========== 用户管理 ==========

// UserList 获取用户列表
func (a *AdminHandler) UserList(c *gin.Context) {
	req := new(request.AdminUserListRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUserList(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// CreateUser 创建用户
func (a *AdminHandler) CreateUser(c *gin.Context) {
	req := new(request.AdminCreateUserRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminCreateUser(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// UpdateUser 更新用户
func (a *AdminHandler) UpdateUser(c *gin.Context) {
	req := new(request.AdminUpdateUserRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUpdateUser(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// DeleteUser 删除用户
func (a *AdminHandler) DeleteUser(c *gin.Context) {
	req := new(request.AdminDeleteUserRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminDeleteUser(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// ToggleUserState 启用/禁用用户
func (a *AdminHandler) ToggleUserState(c *gin.Context) {
	req := new(request.AdminToggleUserStateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminToggleUserState(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// ========== 组管理 ==========

// GroupList 获取组列表
func (a *AdminHandler) GroupList(c *gin.Context) {
	res, err := a.service.AdminGroupList()
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// CreateGroup 创建组
func (a *AdminHandler) CreateGroup(c *gin.Context) {
	req := new(request.AdminCreateGroupRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminCreateGroup(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// UpdateGroup 更新组
func (a *AdminHandler) UpdateGroup(c *gin.Context) {
	req := new(request.AdminUpdateGroupRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUpdateGroup(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// DeleteGroup 删除组
func (a *AdminHandler) DeleteGroup(c *gin.Context) {
	req := new(request.AdminDeleteGroupRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminDeleteGroup(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// ========== 权限管理 ==========

// PowerList 获取权限列表
func (a *AdminHandler) PowerList(c *gin.Context) {
	res, err := a.service.AdminPowerList()
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// AssignPower 分配权限
func (a *AdminHandler) AssignPower(c *gin.Context) {
	req := new(request.AdminAssignPowerRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminAssignPower(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// GetGroupPowers 获取组的权限列表
func (a *AdminHandler) GetGroupPowers(c *gin.Context) {
	req := new(request.AdminGetGroupPowersRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminGetGroupPowers(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// CreatePower 创建权限
func (a *AdminHandler) CreatePower(c *gin.Context) {
	req := new(request.AdminCreatePowerRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminCreatePower(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// UpdatePower 更新权限
func (a *AdminHandler) UpdatePower(c *gin.Context) {
	req := new(request.AdminUpdatePowerRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUpdatePower(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// DeletePower 删除权限
func (a *AdminHandler) DeletePower(c *gin.Context) {
	req := new(request.AdminDeletePowerRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminDeletePower(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// BatchDeletePower 批量删除权限
func (a *AdminHandler) BatchDeletePower(c *gin.Context) {
	req := new(request.AdminBatchDeletePowerRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminBatchDeletePower(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// ========== 磁盘管理 ==========

// DiskList 获取磁盘列表
func (a *AdminHandler) DiskList(c *gin.Context) {
	res, err := a.service.AdminDiskList()
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// CreateDisk 创建磁盘
func (a *AdminHandler) CreateDisk(c *gin.Context) {
	req := new(request.AdminCreateDiskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminCreateDisk(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// UpdateDisk 更新磁盘
func (a *AdminHandler) UpdateDisk(c *gin.Context) {
	req := new(request.AdminUpdateDiskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUpdateDisk(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// DeleteDisk 删除磁盘
func (a *AdminHandler) DeleteDisk(c *gin.Context) {
	req := new(request.AdminDeleteDiskRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminDeleteDisk(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// ========== 系统配置 ==========

// GetSystemConfig 获取系统配置
func (a *AdminHandler) GetSystemConfig(c *gin.Context) {
	res, err := a.service.AdminGetSystemConfig()
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

// UpdateSystemConfig 更新系统配置
func (a *AdminHandler) UpdateSystemConfig(c *gin.Context) {
	req := new(request.AdminUpdateSystemConfigRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, models.NewJsonResponse(400, "参数错误", nil))
		return
	}
	res, err := a.service.AdminUpdateSystemConfig(req)
	if err != nil {
		c.JSON(200, models.NewJsonResponse(400, err.Error(), nil))
		return
	}
	c.JSON(200, res)
}

