package service

import (
	"context"
	"errors"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	factory *impl.RepositoryFactory
}

func NewAdminService(factory *impl.RepositoryFactory) *AdminService {
	return &AdminService{
		factory: factory,
	}
}

func (a *AdminService) GetRepository() *impl.RepositoryFactory {
	return a.factory
}

// ========== 用户管理 ==========

// AdminUserList 获取用户列表
func (a *AdminService) AdminUserList(req *request.AdminUserListRequest) (*models.JsonResponse, error) {
	ctx := context.Background()
	offset := (req.Page - 1) * req.PageSize

	var users []*models.UserInfo
	var total int64
	var err error

	// 使用数据库查询构建器
	db := a.factory.DB()
	query := db.WithContext(ctx).Model(&models.UserInfo{})

	// 关键词搜索
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("user_name LIKE ? OR name LIKE ? OR email LIKE ?", keyword, keyword, keyword)
	}

	// 组筛选
	if req.GroupID > 0 {
		query = query.Where("group_id = ?", req.GroupID)
	}

	// 状态筛选（使用指针类型以区分未传递和传递了0）
	if req.State != nil && *req.State >= 0 {
		query = query.Where("state = ?", *req.State)
	}

	// 获取总数
	if err = query.Count(&total).Error; err != nil {
		logger.LOG.Error("统计用户数量失败", "error", err)
		return nil, err
	}

	// 获取列表
	if err = query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		logger.LOG.Error("查询用户列表失败", "error", err)
		return nil, err
	}

	// 填充组名
	userInfos := make([]*response.AdminUserInfo, 0, len(users))
	for _, user := range users {
		group, err := a.factory.Group().GetByID(ctx, user.GroupID)
		groupName := ""
		if err == nil && group != nil {
			groupName = group.Name
		}
		userInfos = append(userInfos, &response.AdminUserInfo{
			UserInfo:  *user,
			GroupName: groupName,
		})
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminUserListResponse{
		Users:    userInfos,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}), nil
}

// AdminCreateUser 创建用户
func (a *AdminService) AdminCreateUser(req *request.AdminCreateUserRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查用户名是否已存在
	existingUser, err := a.factory.User().GetByUserName(ctx, req.UserName)
	if err != nil && err != gorm.ErrRecordNotFound {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 检查组是否存在
	group, err := a.factory.Group().GetByID(ctx, req.GroupID)
	if err != nil {
		logger.LOG.Error("查询组失败", "error", err)
		return nil, fmt.Errorf("用户组不存在")
	}

	// 检查用户存储空间是否超过组存储空间限制
	// 如果组有存储空间限制（group.Space > 0），则用户存储空间不能超过组存储空间
	if group.Space > 0 && req.Space > 0 && req.Space > group.Space {
		return nil, fmt.Errorf("用户存储空间不能超过组存储空间限制（组限制：%d 字节）", group.Space)
	}

	// 生成密码哈希
	password, err := util.GeneratePassword(req.Password)
	if err != nil {
		logger.LOG.Error("生成密码失败", "error", err)
		return nil, err
	}

	// 创建用户
	v7, err := uuid.NewV7()
	if err != nil {
		logger.LOG.Error("生成UUID失败", "error", err)
		return nil, err
	}

	user := &models.UserInfo{
		ID:        v7.String(),
		Name:      req.Name,
		UserName:  req.UserName,
		Password:  password,
		Email:     req.Email,
		Phone:     req.Phone,
		GroupID:   req.GroupID,
		Space:     req.Space,
		FreeSpace: req.Space,
		CreatedAt: custom_type.Now(),
		State:     0,
	}
	if req.Space == 0 && group.Space > 0 {
		user.Space = group.Space * 1024 * 1024 * 1024
		user.FreeSpace = group.Space * 1024 * 1024 * 1024 // Convert to bytes (GB)
	}

	if err = a.factory.User().Create(ctx, user); err != nil {
		logger.LOG.Error("创建用户失败", "error", err)
		return nil, err
	}
	if err := a.factory.VirtualPath().Create(ctx, &models.VirtualPath{
		UserID:      user.ID,
		Path:        "home",
		CreatedTime: custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}); err != nil {
		logger.LOG.Error("创建虚拟路径失败", "error", err)
		a.factory.User().Delete(ctx, user.ID)
		return nil, err
	}
	return models.NewJsonResponse(200, "创建成功", user), nil
}

// AdminUpdateUser 更新用户
func (a *AdminService) AdminUpdateUser(req *request.AdminUpdateUserRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 获取用户
	user, err := a.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, fmt.Errorf("用户不存在")
	}

	// 更新字段
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.GroupID > 0 {
		// 检查组是否存在
		group, err := a.factory.Group().GetByID(ctx, req.GroupID)
		if err != nil {
			logger.LOG.Error("查询组失败", "error", err)
			return nil, fmt.Errorf("用户组不存在")
		}
		user.GroupID = req.GroupID
		// 如果更新了组，可能需要更新存储空间
		if req.Space == 0 && group.Space > 0 {
			user.Space = group.Space
			user.FreeSpace = group.Space
		}
		// 检查用户存储空间是否超过组存储空间限制
		// 如果组有存储空间限制（group.Space > 0），且用户设置了存储空间（req.Space > 0），则不能超过组限制
		if group.Space > 0 && req.Space > 0 && req.Space > group.Space {
			return nil, fmt.Errorf("用户存储空间不能超过组存储空间限制（组限制：%d 字节）", group.Space)
		}
	}
	if req.Space > 0 {
		// 如果用户组有存储空间限制，需要再次检查（因为可能只更新了存储空间，没有更新组）
		if user.GroupID > 0 {
			group, err := a.factory.Group().GetByID(ctx, user.GroupID)
			if err == nil && group != nil && group.Space > 0 && req.Space > group.Space {
				return nil, fmt.Errorf("用户存储空间不能超过组存储空间限制（组限制：%d 字节）", group.Space)
			}
		}
		user.Space = req.Space
		// 调整剩余空间
		used := user.Space - user.FreeSpace
		if used < 0 {
			used = 0
		}
		user.FreeSpace = req.Space - used
	}
	if req.State >= 0 {
		user.State = req.State
	}

	if err = a.factory.User().Update(ctx, user); err != nil {
		logger.LOG.Error("更新用户失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "更新成功", user), nil
}

// AdminDeleteUser 删除用户
func (a *AdminService) AdminDeleteUser(req *request.AdminDeleteUserRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查用户是否存在
	user, err := a.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, fmt.Errorf("用户不存在")
	}

	// 不能删除管理员（group_id = 1）
	if user.GroupID == 1 {
		return nil, fmt.Errorf("不能删除管理员用户")
	}

	if err = a.factory.User().Delete(ctx, req.ID); err != nil {
		logger.LOG.Error("删除用户失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// AdminToggleUserState 启用/禁用用户
func (a *AdminService) AdminToggleUserState(req *request.AdminToggleUserStateRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	user, err := a.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, fmt.Errorf("用户不存在")
	}

	user.State = req.State

	if err = a.factory.User().Update(ctx, user); err != nil {
		logger.LOG.Error("更新用户状态失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "操作成功", user), nil
}

// ========== 组管理 ==========

// AdminGroupList 获取组列表
func (a *AdminService) AdminGroupList() (*models.JsonResponse, error) {
	ctx := context.Background()

	groups, err := a.factory.Group().List(ctx, 0, 1000) // 获取所有组
	if err != nil {
		logger.LOG.Error("查询组列表失败", "error", err)
		return nil, err
	}

	total, err := a.factory.Group().Count(ctx)
	if err != nil {
		logger.LOG.Error("统计组数量失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminGroupListResponse{
		Groups: groups,
		Total:  total,
	}), nil
}

// AdminCreateGroup 创建组
func (a *AdminService) AdminCreateGroup(req *request.AdminCreateGroupRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查组名是否已存在
	groups, err := a.factory.Group().List(ctx, 0, 1000)
	if err != nil {
		logger.LOG.Error("查询组列表失败", "error", err)
		return nil, err
	}
	for _, g := range groups {
		if g.Name == req.Name {
			return nil, fmt.Errorf("组名已存在")
		}
	}

	// 获取最大ID
	maxID := 0
	for _, g := range groups {
		if g.ID > maxID {
			maxID = g.ID
		}
	}

	group := &models.Group{
		ID:           maxID + 1,
		Name:         req.Name,
		GroupDefault: req.GroupDefault,
		Space:        req.Space,
		CreatedAt:    custom_type.Now(),
	}

	if err = a.factory.Group().Create(ctx, group); err != nil {
		logger.LOG.Error("创建组失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "创建成功", group), nil
}

// AdminUpdateGroup 更新组
func (a *AdminService) AdminUpdateGroup(req *request.AdminUpdateGroupRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	group, err := a.factory.Group().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询组失败", "error", err)
		return nil, fmt.Errorf("组不存在")
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Space >= 0 {
		group.Space = req.Space
	}
	if req.GroupDefault >= 0 {
		group.GroupDefault = req.GroupDefault
	}

	if err = a.factory.Group().Update(ctx, group); err != nil {
		logger.LOG.Error("更新组失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "更新成功", group), nil
}

// AdminDeleteGroup 删除组
func (a *AdminService) AdminDeleteGroup(req *request.AdminDeleteGroupRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 不能删除管理员组（ID = 1）
	if req.ID == 1 {
		return nil, fmt.Errorf("不能删除管理员组")
	}

	// 检查是否有用户使用该组
	users, err := a.factory.User().List(ctx, 0, 1000)
	if err == nil {
		for _, user := range users {
			if user.GroupID == req.ID {
				return nil, fmt.Errorf("该组下还有用户，无法删除")
			}
		}
	}

	if err = a.factory.Group().Delete(ctx, req.ID); err != nil {
		logger.LOG.Error("删除组失败", "error", err)
		return nil, err
	}

	// 删除组的权限关联
	if err = a.factory.GroupPower().DeleteByGroupID(ctx, req.ID); err != nil {
		logger.LOG.Warn("删除组权限关联失败", "error", err)
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// ========== 权限管理 ==========

// AdminPowerList 获取权限列表
func (a *AdminService) AdminPowerList(req *request.AdminPowerListRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 获取总数
	total, err := a.factory.Power().Count(ctx)
	if err != nil {
		logger.LOG.Error("统计权限数量失败", "error", err)
		return nil, err
	}

	// 获取列表
	powers, err := a.factory.Power().List(ctx, offset, pageSize)
	if err != nil {
		logger.LOG.Error("查询权限列表失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminPowerListResponse{
		Powers: powers,
		Total:  total,
	}), nil
}

// AdminAssignPower 为组分配权限
func (a *AdminService) AdminAssignPower(req *request.AdminAssignPowerRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查组是否存在
	_, err := a.factory.Group().GetByID(ctx, req.GroupID)
	if err != nil {
		logger.LOG.Error("查询组失败", "error", err)
		return nil, fmt.Errorf("组不存在")
	}

	// 删除原有权限
	if err = a.factory.GroupPower().DeleteByGroupID(ctx, req.GroupID); err != nil {
		logger.LOG.Warn("删除组权限失败", "error", err)
	}

	// 创建新权限关联
	groupPowers := make([]*models.GroupPower, 0, len(req.PowerIDs))
	for _, powerID := range req.PowerIDs {
		// 检查权限是否存在
		_, err := a.factory.Power().GetByID(ctx, powerID)
		if err != nil {
			logger.LOG.Warn("权限不存在，跳过", "power_id", powerID)
			continue
		}
		groupPowers = append(groupPowers, &models.GroupPower{
			GroupID: req.GroupID,
			PowerID: powerID,
		})
	}

	if len(groupPowers) > 0 {
		if err = a.factory.GroupPower().BatchCreate(ctx, groupPowers); err != nil {
			logger.LOG.Error("分配权限失败", "error", err)
			return nil, err
		}
	}

	return models.NewJsonResponse(200, "分配成功", nil), nil
}

// AdminGetGroupPowers 获取组的权限列表
func (a *AdminService) AdminGetGroupPowers(req *request.AdminGetGroupPowersRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	groupPowers, err := a.factory.GroupPower().GetByGroupID(ctx, req.GroupID)
	if err != nil {
		logger.LOG.Error("查询组权限失败", "error", err)
		return nil, err
	}

	powerIDs := make([]int, 0, len(groupPowers))
	for _, gp := range groupPowers {
		powerIDs = append(powerIDs, gp.PowerID)
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminGroupPowersResponse{
		PowerIDs: powerIDs,
	}), nil
}

// AdminCreatePower 创建权限
func (a *AdminService) AdminCreatePower(req *request.AdminCreatePowerRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	power := &models.Power{
		Name:           req.Name,
		Description:    req.Description,
		Characteristic: req.Characteristic,
	}

	if err := a.factory.Power().Create(ctx, power); err != nil {
		logger.LOG.Error("创建权限失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "创建成功", power), nil
}

// AdminUpdatePower 更新权限
func (a *AdminService) AdminUpdatePower(req *request.AdminUpdatePowerRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查权限是否存在
	power, err := a.factory.Power().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询权限失败", "error", err)
		return nil, fmt.Errorf("权限不存在")
	}

	// 更新字段
	if req.Name != "" {
		power.Name = req.Name
	}
	if req.Description != "" {
		power.Description = req.Description
	}
	if req.Characteristic != "" {
		power.Characteristic = req.Characteristic
	}

	if err = a.factory.Power().Update(ctx, power); err != nil {
		logger.LOG.Error("更新权限失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "更新成功", power), nil
}

// AdminDeletePower 删除权限
func (a *AdminService) AdminDeletePower(req *request.AdminDeletePowerRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查权限是否存在
	_, err := a.factory.Power().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询权限失败", "error", err)
		return nil, fmt.Errorf("权限不存在")
	}

	// 检查是否有组使用该权限
	groupPowers, err := a.factory.GroupPower().GetByPowerID(ctx, req.ID)
	if err == nil && len(groupPowers) > 0 {
		return nil, fmt.Errorf("该权限正在被使用，无法删除")
	}

	if err = a.factory.Power().Delete(ctx, req.ID); err != nil {
		logger.LOG.Error("删除权限失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// AdminBatchDeletePower 批量删除权限
func (a *AdminService) AdminBatchDeletePower(req *request.AdminBatchDeletePowerRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	var successCount int
	var failedItems []string
	var inUseItems []string

	for _, id := range req.IDs {
		// 检查权限是否存在
		power, err := a.factory.Power().GetByID(ctx, id)
		if err != nil {
			logger.LOG.Warn("查询权限失败", "power_id", id, "error", err)
			failedItems = append(failedItems, fmt.Sprintf("ID:%d(不存在)", id))
			continue
		}

		// 检查是否有组使用该权限
		groupPowers, err := a.factory.GroupPower().GetByPowerID(ctx, id)
		if err == nil && len(groupPowers) > 0 {
			inUseItems = append(inUseItems, fmt.Sprintf("%s(ID:%d)", power.Name, id))
			continue
		}

		// 删除权限
		if err = a.factory.Power().Delete(ctx, id); err != nil {
			logger.LOG.Error("删除权限失败", "power_id", id, "error", err)
			failedItems = append(failedItems, fmt.Sprintf("%s(ID:%d)", power.Name, id))
			continue
		}

		successCount++
	}

	// 构建返回消息
	var message string
	if successCount == len(req.IDs) {
		message = fmt.Sprintf("成功删除 %d 个权限", successCount)
	} else {
		message = fmt.Sprintf("成功删除 %d 个权限", successCount)
		if len(inUseItems) > 0 {
			message += fmt.Sprintf("，%d 个权限正在被使用无法删除：%s", len(inUseItems), fmt.Sprintf("%v", inUseItems))
		}
		if len(failedItems) > 0 {
			message += fmt.Sprintf("，%d 个权限删除失败：%s", len(failedItems), fmt.Sprintf("%v", failedItems))
		}
	}

	return models.NewJsonResponse(200, message, map[string]interface{}{
		"success_count": successCount,
		"total_count":   len(req.IDs),
		"in_use_items":  inUseItems,
		"failed_items":  failedItems,
	}), nil
}

// ========== 磁盘管理 ==========

// AdminDiskList 获取磁盘列表
func (a *AdminService) AdminDiskList() (*models.JsonResponse, error) {
	ctx := context.Background()

	disks, err := a.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		logger.LOG.Error("查询磁盘列表失败", "error", err)
		return nil, err
	}

	total, err := a.factory.Disk().Count(ctx)
	if err != nil {
		logger.LOG.Error("统计磁盘数量失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminDiskListResponse{
		Disks: disks,
		Total: total,
	}), nil
}

// AdminCreateDisk 创建磁盘
func (a *AdminService) AdminCreateDisk(req *request.AdminCreateDiskRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查路径是否已存在
	existingDisk, err := a.factory.Disk().GetByPath(ctx, req.DiskPath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询磁盘失败", "error", err)
		return nil, err
	}
	if existingDisk != nil {
		return nil, fmt.Errorf("磁盘路径已存在")
	}

	// 生成磁盘ID
	diskID := uuid.New().String()

	disk := &models.Disk{
		ID:       diskID,
		DiskPath: req.DiskPath,
		DataPath: req.DataPath,
		Size:     req.Size,
	}

	if err = a.factory.Disk().Create(ctx, disk); err != nil {
		logger.LOG.Error("创建磁盘失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "创建成功", disk), nil
}

// AdminUpdateDisk 更新磁盘
func (a *AdminService) AdminUpdateDisk(req *request.AdminUpdateDiskRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	disk, err := a.factory.Disk().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询磁盘失败", "error", err)
		return nil, fmt.Errorf("磁盘不存在")
	}

	if req.DiskPath != "" {
		disk.DiskPath = req.DiskPath
	}
	if req.DataPath != "" {
		disk.DataPath = req.DataPath
	}
	if req.Size > 0 {
		disk.Size = req.Size * 1024 * 1024 * 1024 // GB转字节
	}

	if err = a.factory.Disk().Update(ctx, disk); err != nil {
		logger.LOG.Error("更新磁盘失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "更新成功", disk), nil
}

// AdminDeleteDisk 删除磁盘
func (a *AdminService) AdminDeleteDisk(req *request.AdminDeleteDiskRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查磁盘是否存在
	_, err := a.factory.Disk().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询磁盘失败", "error", err)
		return nil, fmt.Errorf("磁盘不存在")
	}

	// TODO: 检查是否有文件存储在该磁盘上

	if err = a.factory.Disk().Delete(ctx, req.ID); err != nil {
		logger.LOG.Error("删除磁盘失败", "error", err)
		return nil, err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// ========== 系统配置 ==========

// AdminGetSystemConfig 获取系统配置
func (a *AdminService) AdminGetSystemConfig() (*models.JsonResponse, error) {
	ctx := context.Background()

	// 获取配置
	allowRegister, _ := a.factory.SysConfig().GetByKey(ctx, "allow_register")
	webdavEnabled, _ := a.factory.SysConfig().GetByKey(ctx, "webdav_enabled")

	// 获取统计信息
	totalUsers, _ := a.factory.User().Count(ctx)
	totalFiles, _ := a.factory.FileInfo().Count(ctx)
	config := response.AdminSystemConfigResponse{
		AllowRegister: allowRegister != nil && allowRegister.Value == "true",
		WebdavEnabled: webdavEnabled != nil && webdavEnabled.Value == "true",
		Version:       "1.0.0", // TODO: 从配置或构建信息获取
		TotalUsers:    totalUsers,
		TotalFiles:    totalFiles,
		Uptime:        custom_type.GetSystemRuntime().String(),
	}

	return models.NewJsonResponse(200, "查询成功", config), nil
}

// AdminUpdateSystemConfig 更新系统配置
func (a *AdminService) AdminUpdateSystemConfig(req *request.AdminUpdateSystemConfigRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	configs := make([]*models.SysConfig, 0)

	if req.AllowRegister {
		config, _ := a.factory.SysConfig().GetByKey(ctx, "allow_register")
		if config == nil {
			config = &models.SysConfig{Key: "allow_register", Value: "true"}
		} else {
			config.Value = "true"
		}
		configs = append(configs, config)
	} else {
		config, _ := a.factory.SysConfig().GetByKey(ctx, "allow_register")
		if config == nil {
			config = &models.SysConfig{Key: "allow_register", Value: "false"}
		} else {
			config.Value = "false"
		}
		configs = append(configs, config)
	}

	if req.WebdavEnabled {
		config, _ := a.factory.SysConfig().GetByKey(ctx, "webdav_enabled")
		if config == nil {
			config = &models.SysConfig{Key: "webdav_enabled", Value: "true"}
		} else {
			config.Value = "true"
		}
		configs = append(configs, config)
	} else {
		config, _ := a.factory.SysConfig().GetByKey(ctx, "webdav_enabled")
		if config == nil {
			config = &models.SysConfig{Key: "webdav_enabled", Value: "false"}
		} else {
			config.Value = "false"
		}
		configs = append(configs, config)
	}

	// 批量更新
	for _, cfg := range configs {
		if cfg.ID == 0 {
			if err := a.factory.SysConfig().Create(ctx, cfg); err != nil {
				logger.LOG.Error("创建配置失败", "key", cfg.Key, "error", err)
			}
		} else {
			if err := a.factory.SysConfig().Update(ctx, cfg); err != nil {
				logger.LOG.Error("更新配置失败", "key", cfg.Key, "error", err)
			}
		}
	}

	return a.AdminGetSystemConfig()
}
