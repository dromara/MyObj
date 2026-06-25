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
	"path/filepath"
	"time"

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
// 所有 AdminService 方法使用 30 秒超时的 context，防止数据库响应缓慢时请求无限等待。
func (a *AdminService) AdminUserList(ctx context.Context, req *request.AdminUserListRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	offset := (req.Page - 1) * req.PageSize

	var users []*models.UserInfo
	var total int64
	var err error

	// 使用数据库查询构建器
	db := a.factory.DB()
	query := db.WithContext(ctx).Model(&models.UserInfo{})

	// 关键词搜索
	if req.Keyword != "" {
		safeKeyword := util.EscapeLikeKeyword(req.Keyword)
		keyword := "%" + safeKeyword + "%"
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
		return nil, fmt.Errorf("统计用户数量失败: %w", err)
	}

	// 获取列表
	if err = query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		logger.LOG.Error("查询用户列表失败", "error", err)
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	// 填充组名（批量查询避免 N+1）
	groupIDSet := make(map[int]struct{})
	for _, user := range users {
		if user.GroupID > 0 {
			groupIDSet[user.GroupID] = struct{}{}
		}
	}
	groupNameMap := make(map[int]string, len(groupIDSet))
	for groupID := range groupIDSet {
		group, err := a.factory.Group().GetByID(ctx, groupID)
		if err == nil && group != nil {
			groupNameMap[groupID] = group.Name
		}
	}

	userInfos := make([]*response.AdminUserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, &response.AdminUserInfo{
			UserInfo:  *user,
			GroupName: groupNameMap[user.GroupID],
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
func (a *AdminService) AdminCreateUser(ctx context.Context, req *request.AdminCreateUserRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 检查用户名是否已存在
	existingUser, err := a.factory.User().GetByUserName(ctx, req.UserName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, fmt.Errorf("查询用户失败: %w", err)
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
		return nil, fmt.Errorf("生成密码失败: %w", err)
	}

	// 创建用户
	v7, err := uuid.NewV7()
	if err != nil {
		logger.LOG.Error("生成UUID失败", "error", err)
		return nil, fmt.Errorf("生成UUID失败: %w", err)
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
	// 如果用户Space为0且组有限制，使用组的Space
	if req.Space == 0 && group.Space > 0 {
		user.Space = group.Space
		user.FreeSpace = group.Space
	}

	// 使用事务包裹 User + VirtualPath 创建，保证原子性
	tx := a.factory.DB().Begin()
	if tx.Error != nil {
		logger.LOG.Error("开启事务失败", "error", tx.Error)
		return nil, fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	txF := a.factory.WithTx(tx)
	if err = txF.User().Create(ctx, user); err != nil {
		tx.Rollback()
		logger.LOG.Error("创建用户失败", "error", err)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	if err := txF.VirtualPath().Create(ctx, &models.VirtualPath{
		UserID:      user.ID,
		Path:        "home",
		CreatedTime: custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}); err != nil {
		tx.Rollback()
		logger.LOG.Error("创建虚拟路径失败", "error", err)
		return nil, fmt.Errorf("创建虚拟路径失败: %w", err)
	}
	if err = tx.Commit().Error; err != nil {
		logger.LOG.Error("提交事务失败", "error", err)
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}
	return models.NewJsonResponse(200, "创建成功", user), nil
}

// AdminUpdateUser 更新用户
func (a *AdminService) AdminUpdateUser(ctx context.Context, req *request.AdminUpdateUserRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		// 仅当管理员显式指定 Space 时才覆盖用户空间，避免修改组时意外重置空间
		// 检查用户存储空间是否超过组存储空间限制
		// 如果组有存储空间限制（group.Space > 0），且用户设置了存储空间（req.Space > 0），则不能超过组限制
		if group.Space > 0 && req.Space > 0 && req.Space > group.Space {
			return nil, fmt.Errorf("用户存储空间不能超过组存储空间限制（组限制：%d 字节）", group.Space)
		}
	}
	// req.Space == 0 表示不修改用户存储空间（保持原值），而非设为"无限"。
	// 如需支持"无限空间"语义，应使用 *int64 指针类型区分"未传递"和"传递了 0"。
	if req.Space > 0 {
		// 如果用户组有存储空间限制，需要再次检查（因为可能只更新了存储空间，没有更新组）
		if user.GroupID > 0 {
			group, err := a.factory.Group().GetByID(ctx, user.GroupID)
			if err == nil && group != nil && group.Space > 0 && req.Space > group.Space {
				return nil, fmt.Errorf("用户存储空间不能超过组存储空间限制（组限制：%d 字节）", group.Space)
			}
		}
		user.Space = req.Space
		// 调整剩余空间：已用空间 = 旧总空间 - 旧剩余空间
		used := user.Space - user.FreeSpace
		if used < 0 {
			used = 0
		}
		user.FreeSpace = req.Space - used
		// 当 FreeSpace 为负数时（新总空间 < 已用空间），将其归零以避免无效的负值
		if user.FreeSpace < 0 {
			user.FreeSpace = 0
		}
	}
	if req.State != nil {
		user.State = *req.State
	}

	if err = a.factory.User().Update(ctx, user); err != nil {
		logger.LOG.Error("更新用户失败", "error", err)
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return models.NewJsonResponse(200, "更新成功", user), nil
}

// AdminDeleteUser 删除用户（在同一事务中级联清理所有关联数据）
func (a *AdminService) AdminDeleteUser(ctx context.Context, req *request.AdminDeleteUserRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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

	err = a.factory.DB().Transaction(func(tx *gorm.DB) error {
		txF := a.factory.WithTx(tx)
		// 1. 删除用户的 API Keys
		if err := txF.ApiKey().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户API Keys失败: %w", err)
		}
		// 2. 删除用户的分享链接
		if err := txF.Share().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户分享链接失败: %w", err)
		}
		// 3. 删除用户的下载任务
		if err := txF.DownloadTask().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户下载任务失败: %w", err)
		}
		// 4. 删除用户的上传任务
		if err := txF.UploadTask().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户上传任务失败: %w", err)
		}
		// 5. 删除回收站记录
		if err := txF.Recycled().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户回收站记录失败: %w", err)
		}
		// 6. 删除用户文件关联
		if err := txF.UserFiles().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户文件关联失败: %w", err)
		}
		// 7. 删除虚拟路径
		if err := txF.VirtualPath().DeleteByUserID(ctx, req.ID); err != nil {
			return fmt.Errorf("删除用户虚拟路径失败: %w", err)
		}
		// 8. 删除用户
		return txF.User().Delete(ctx, req.ID)
	})
	if err != nil {
		logger.LOG.Error("删除用户失败", "error", err)
		return nil, fmt.Errorf("删除用户失败: %w", err)
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// AdminToggleUserState 启用/禁用用户
func (a *AdminService) AdminToggleUserState(ctx context.Context, req *request.AdminToggleUserStateRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	user, err := a.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, fmt.Errorf("用户不存在")
	}

	user.State = req.State

	if err = a.factory.User().Update(ctx, user); err != nil {
		logger.LOG.Error("更新用户状态失败", "error", err)
		return nil, fmt.Errorf("更新用户状态失败: %w", err)
	}

	return models.NewJsonResponse(200, "操作成功", user), nil
}

// ========== 组管理 ==========

// AdminGroupList 获取组列表
func (a *AdminService) AdminGroupList(ctx context.Context) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 使用较大的 limit 一次性查询，避免 Count+List 之间的竞态
	const maxLimit = 10000
	groups, err := a.factory.Group().List(ctx, 0, maxLimit)
	if err != nil {
		logger.LOG.Error("查询组列表失败", "error", err)
		return nil, fmt.Errorf("查询组列表失败: %w", err)
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminGroupListResponse{
		Groups: groups,
		Total:  int64(len(groups)),
	}), nil
}

// AdminCreateGroup 创建组
func (a *AdminService) AdminCreateGroup(ctx context.Context, req *request.AdminCreateGroupRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 使用事务保证"查询最大ID + 创建"的原子性，避免并发时产生重复ID
	var group *models.Group
	err := a.factory.DB().Transaction(func(tx *gorm.DB) error {
		txF := a.factory.WithTx(tx)

		// 检查组名是否已存在（加载全部组后过滤，因为 GroupRepository 暂无 GetByName）
		total, err := txF.Group().Count(ctx)
		if err != nil {
			return fmt.Errorf("统计组数量失败: %w", err)
		}
		existingGroups, err := txF.Group().List(ctx, 0, int(total)+1)
		if err != nil {
			return fmt.Errorf("查询组列表失败: %w", err)
		}
		for _, g := range existingGroups {
			if g.Name == req.Name {
				return fmt.Errorf("组名已存在")
			}
		}

		// 在事务内查询最大ID，保证并发安全
		maxID, err := txF.Group().GetMaxID(ctx)
		if err != nil {
			return fmt.Errorf("查询最大组ID失败: %w", err)
		}

		group = &models.Group{
			ID:           maxID + 1,
			Name:         req.Name,
			GroupDefault: req.GroupDefault,
			Space:        req.Space,
			CreatedAt:    custom_type.Now(),
		}

		if err = txF.Group().Create(ctx, group); err != nil {
			return fmt.Errorf("创建组失败: %w", err)
		}
		return nil
	})
	if err != nil {
		logger.LOG.Error("创建组失败", "error", err)
		return nil, fmt.Errorf("创建组失败: %w", err)
	}

	return models.NewJsonResponse(200, "创建成功", group), nil
}

// AdminUpdateGroup 更新组
func (a *AdminService) AdminUpdateGroup(ctx context.Context, req *request.AdminUpdateGroupRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		return nil, fmt.Errorf("更新组失败: %w", err)
	}

	return models.NewJsonResponse(200, "更新成功", group), nil
}

// AdminDeleteGroup 删除组
func (a *AdminService) AdminDeleteGroup(ctx context.Context, req *request.AdminDeleteGroupRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 不能删除管理员组（ID = 1）
	if req.ID == 1 {
		return nil, fmt.Errorf("不能删除管理员组")
	}

	// 检查是否有用户使用该组（精确查询，避免加载全部用户）
	userCount, err := a.factory.User().CountByGroupID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("统计组用户数量失败", "error", err)
		return nil, fmt.Errorf("统计组用户数量失败: %w", err)
	}
	if userCount > 0 {
		return nil, fmt.Errorf("该组下还有用户，无法删除")
	}

	// 在同一个事务中先删权限关联再删组，保证原子性
	tx := a.factory.DB().Begin()
	if tx.Error != nil {
		logger.LOG.Error("开启事务失败", "error", tx.Error)
		return nil, fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	txF := a.factory.WithTx(tx)
	if err = txF.GroupPower().DeleteByGroupID(ctx, req.ID); err != nil {
		tx.Rollback()
		logger.LOG.Error("删除组权限关联失败", "error", err)
		return nil, fmt.Errorf("删除组权限关联失败: %w", err)
	}
	if err = txF.Group().Delete(ctx, req.ID); err != nil {
		tx.Rollback()
		logger.LOG.Error("删除组失败", "error", err)
		return nil, fmt.Errorf("删除组失败: %w", err)
	}
	if err = tx.Commit().Error; err != nil {
		logger.LOG.Error("提交事务失败", "error", err)
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// ========== 权限管理 ==========

// AdminPowerList 获取权限列表
func (a *AdminService) AdminPowerList(ctx context.Context, req *request.AdminPowerListRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	// 允许更大的 pageSize（最大 1000），用于组管理分配权限时获取所有权限
	if pageSize > 1000 {
		pageSize = 1000
	}
	offset := (page - 1) * pageSize

	// 获取总数
	total, err := a.factory.Power().Count(ctx)
	if err != nil {
		logger.LOG.Error("统计权限数量失败", "error", err)
		return nil, fmt.Errorf("统计权限数量失败: %w", err)
	}

	// 获取列表
	powers, err := a.factory.Power().List(ctx, offset, pageSize)
	if err != nil {
		logger.LOG.Error("查询权限列表失败", "error", err)
		return nil, fmt.Errorf("查询权限列表失败: %w", err)
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminPowerListResponse{
		Powers: powers,
		Total:  total,
	}), nil
}

// AdminAssignPower 为组分配权限
func (a *AdminService) AdminAssignPower(ctx context.Context, req *request.AdminAssignPowerRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 检查组是否存在
	_, err := a.factory.Group().GetByID(ctx, req.GroupID)
	if err != nil {
		logger.LOG.Error("查询组失败", "error", err)
		return nil, fmt.Errorf("组不存在")
	}

	// 使用事务包裹"删除旧权限 + 创建新权限"，避免中间失败导致权限丢失
	err = a.factory.DB().Transaction(func(tx *gorm.DB) error {
		txF := a.factory.WithTx(tx)

		// 删除原有权限
		if err := txF.GroupPower().DeleteByGroupID(ctx, req.GroupID); err != nil {
			return fmt.Errorf("删除组权限失败: %w", err)
		}

		// 检查权限是否存在并收集有效的权限ID
		validPowerIDs := make(map[int]struct{}, len(req.PowerIDs))
		for _, powerID := range req.PowerIDs {
			if _, err := txF.Power().GetByID(ctx, powerID); err != nil {
				logger.LOG.Warn("权限不存在，跳过", "power_id", powerID)
				continue
			}
			validPowerIDs[powerID] = struct{}{}
		}

		// 创建新权限关联
		groupPowers := make([]*models.GroupPower, 0, len(validPowerIDs))
		for powerID := range validPowerIDs {
			groupPowers = append(groupPowers, &models.GroupPower{
				GroupID: req.GroupID,
				PowerID: powerID,
			})
		}

		if len(groupPowers) > 0 {
			if err := txF.GroupPower().BatchCreate(ctx, groupPowers); err != nil {
				return fmt.Errorf("分配权限失败: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		logger.LOG.Error("分配权限失败", "error", err)
		return nil, fmt.Errorf("分配权限失败: %w", err)
	}

	return models.NewJsonResponse(200, "分配成功", nil), nil
}

// AdminGetGroupPowers 获取组的权限列表
func (a *AdminService) AdminGetGroupPowers(ctx context.Context, req *request.AdminGetGroupPowersRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	groupPowers, err := a.factory.GroupPower().GetByGroupID(ctx, req.GroupID)
	if err != nil {
		logger.LOG.Error("查询组权限失败", "error", err)
		return nil, fmt.Errorf("查询组权限失败: %w", err)
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
func (a *AdminService) AdminCreatePower(ctx context.Context, req *request.AdminCreatePowerRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	power := &models.Power{
		Name:           req.Name,
		Description:    req.Description,
		Characteristic: req.Characteristic,
		CreatedAt:      custom_type.Now(),
	}

	if err := a.factory.Power().Create(ctx, power); err != nil {
		logger.LOG.Error("创建权限失败", "error", err)
		return nil, fmt.Errorf("创建权限失败: %w", err)
	}

	return models.NewJsonResponse(200, "创建成功", power), nil
}

// AdminUpdatePower 更新权限
func (a *AdminService) AdminUpdatePower(ctx context.Context, req *request.AdminUpdatePowerRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		return nil, fmt.Errorf("更新权限失败: %w", err)
	}

	return models.NewJsonResponse(200, "更新成功", power), nil
}

// AdminDeletePower 删除权限
func (a *AdminService) AdminDeletePower(ctx context.Context, req *request.AdminDeletePowerRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		return nil, fmt.Errorf("删除权限失败: %w", err)
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// AdminBatchDeletePower 批量删除权限
func (a *AdminService) AdminBatchDeletePower(ctx context.Context, req *request.AdminBatchDeletePowerRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
func (a *AdminService) AdminDiskList(ctx context.Context) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 使用较大的 limit 一次性查询，避免 Count+List 之间的竞态
	const maxLimit = 10000
	disks, err := a.factory.Disk().List(ctx, 0, maxLimit)
	if err != nil {
		logger.LOG.Error("查询磁盘列表失败", "error", err)
		return nil, fmt.Errorf("查询磁盘列表失败: %w", err)
	}

	return models.NewJsonResponse(200, "查询成功", response.AdminDiskListResponse{
		Disks: disks,
		Total: int64(len(disks)),
	}), nil
}

// AdminCreateDisk 创建磁盘
func (a *AdminService) AdminCreateDisk(ctx context.Context, req *request.AdminCreateDiskRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 检查路径是否已存在
	existingDisk, err := a.factory.Disk().GetByPath(ctx, req.DiskPath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询磁盘失败", "error", err)
		return nil, fmt.Errorf("查询磁盘失败: %w", err)
	}
	if existingDisk != nil {
		return nil, fmt.Errorf("磁盘路径已存在")
	}

	// 将DataPath转换为绝对路径（如果是相对路径）
	dataPath := req.DataPath
	if !filepath.IsAbs(dataPath) {
		// 相对路径，转换为绝对路径
		absPath, err := filepath.Abs(dataPath)
		if err != nil {
			logger.LOG.Error("转换绝对路径失败", "error", err, "dataPath", dataPath)
			return nil, fmt.Errorf("无效的数据路径: %w", err)
		}
		dataPath = absPath
		logger.LOG.Info("将相对路径转换为绝对路径", "original", req.DataPath, "absolute", dataPath)
	}

	// 生成磁盘ID
	diskID := uuid.New().String()

	disk := &models.Disk{
		ID:       diskID,
		DiskPath: req.DiskPath,
		DataPath: dataPath, // 使用绝对路径
		Size:     req.Size,
	}

	if err = a.factory.Disk().Create(ctx, disk); err != nil {
		logger.LOG.Error("创建磁盘失败", "error", err)
		return nil, fmt.Errorf("创建磁盘失败: %w", err)
	}

	logger.LOG.Info("磁盘创建成功", "diskID", diskID, "diskPath", req.DiskPath, "dataPath", dataPath)
	return models.NewJsonResponse(200, "创建成功", disk), nil
}

// AdminUpdateDisk 更新磁盘
func (a *AdminService) AdminUpdateDisk(ctx context.Context, req *request.AdminUpdateDiskRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		disk.Size = req.Size // 前端已传递字节
	}

	if err = a.factory.Disk().Update(ctx, disk); err != nil {
		logger.LOG.Error("更新磁盘失败", "error", err)
		return nil, fmt.Errorf("更新磁盘失败: %w", err)
	}

	return models.NewJsonResponse(200, "更新成功", disk), nil
}

// AdminDeleteDisk 删除磁盘
func (a *AdminService) AdminDeleteDisk(ctx context.Context, req *request.AdminDeleteDiskRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 检查磁盘是否存在
	_, err := a.factory.Disk().GetByID(ctx, req.ID)
	if err != nil {
		logger.LOG.Error("查询磁盘失败", "error", err)
		return nil, fmt.Errorf("磁盘不存在")
	}

	// TODO: 检查是否有文件存储在该磁盘上

	if err = a.factory.Disk().Delete(ctx, req.ID); err != nil {
		logger.LOG.Error("删除磁盘失败", "error", err)
		return nil, fmt.Errorf("删除磁盘失败: %w", err)
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// ========== 系统配置 ==========

// AdminGetSystemConfig 获取系统配置
func (a *AdminService) AdminGetSystemConfig(ctx context.Context) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 获取配置
	allowRegister, _ := a.factory.SysConfig().GetByKey(ctx, "allow_register")
	webdavEnabled, _ := a.factory.SysConfig().GetByKey(ctx, "webdav_enabled")

	// 获取统计信息
	totalUsers, _ := a.factory.User().Count(ctx)
	totalFiles, _ := a.factory.FileInfo().Count(ctx)
	config := response.AdminSystemConfigResponse{
		AllowRegister: allowRegister != nil && allowRegister.Value == "true",
		WebdavEnabled: webdavEnabled != nil && webdavEnabled.Value == "true",
		Version:       NewUpgradeService().GetCurrentVersion(),
		TotalUsers:    totalUsers,
		TotalFiles:    totalFiles,
		Uptime:        custom_type.GetSystemRuntime().String(),
	}

	return models.NewJsonResponse(200, "查询成功", config), nil
}

// AdminUpdateSystemConfig 更新系统配置
func (a *AdminService) AdminUpdateSystemConfig(ctx context.Context, req *request.AdminUpdateSystemConfigRequest) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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

	// 使用事务批量更新，保证配置更新的原子性
	tx := a.factory.DB().Begin()
	if tx.Error != nil {
		logger.LOG.Error("开启事务失败", "error", tx.Error)
		return nil, fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	txF := a.factory.WithTx(tx)
	for _, cfg := range configs {
		if cfg.ID == 0 {
			if err := txF.SysConfig().Create(ctx, cfg); err != nil {
				tx.Rollback()
				logger.LOG.Error("创建配置失败", "key", cfg.Key, "error", err)
				return nil, fmt.Errorf("创建配置失败: %w", err)
			}
		} else {
			if err := txF.SysConfig().Update(ctx, cfg); err != nil {
				tx.Rollback()
				logger.LOG.Error("更新配置失败", "key", cfg.Key, "error", err)
				return nil, fmt.Errorf("更新配置失败: %w", err)
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		logger.LOG.Error("提交事务失败", "error", err)
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return a.AdminGetSystemConfig(ctx)
}

// ========== 系统升级 ==========

// CheckUpdate 检查更新
func (a *AdminService) CheckUpdate(ctx context.Context) (*models.JsonResponse, error) {
	upgradeSvc := NewUpgradeService()
	info, err := upgradeSvc.CheckUpdate()
	if err != nil {
		logger.LOG.Error("检查更新失败", "error", err)
		return nil, fmt.Errorf("检查更新失败: %w", err)
	}
	return models.NewJsonResponse(200, "检查成功", info), nil
}

// PerformUpgrade 执行升级
func (a *AdminService) PerformUpgrade(ctx context.Context, downloadURL string) (*models.JsonResponse, error) {
	upgradeSvc := NewUpgradeService()

	// 异步执行升级，先返回响应
	go func() {
		logger.LOG.Info("开始执行系统升级", "url", downloadURL)
		if err := upgradeSvc.PerformUpgrade(downloadURL); err != nil {
			logger.LOG.Error("系统升级失败", "error", err)
			return
		}
		logger.LOG.Info("系统升级脚本已启动，服务即将重启")
	}()

	return models.NewJsonResponse(200, "升级任务已启动，服务即将重启", nil), nil
}
