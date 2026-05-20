package service

import (
	"context"
	"errors"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/auth"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewUserService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *UserService {
	return &UserService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}
func (u *UserService) GetRepository() *impl.RepositoryFactory {
	return u.factory
}

// Login 用户登录
func (u *UserService) Login(username, password, challenge string) (*models.JsonResponse, error) {
	ctx := context.Background()
	get, err := u.cacheLocal.Get(challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, err
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}
	decrypt, err := util.Decrypt(challengeId, password)
	if err != nil {
		logger.LOG.Error("密码挑战验证失败", "error", err)
		return nil, err
	}
	psw := string(decrypt)

	// 验证用户名和密码
	if username == "" || psw == "" {
		return nil, fmt.Errorf("用户名或密码不能为空")
	}
	user, err := u.factory.User().GetByUserName(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
			return nil, fmt.Errorf("用户不存在")
		}
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, err
	}
	if user.State == 1 {
		return nil, fmt.Errorf("用户已被禁用")
	}
	if !util.CheckPassword(user.Password, psw) {
		logger.LOG.Error("密码错误", "error", err)
		return nil, fmt.Errorf("密码错误")
	}
	powers, err := u.factory.Power().GetByGroupID(ctx, user.GroupID)
	if err != nil {
		logger.LOG.Error("查询用户权限失败", "error", err)
		return nil, err
	}
	user.Password = ""
	user.FilePassword = ""
	res := response.UserLoginResponse{
		Token: "",
		User:  user,
		Power: powers,
	}
	uid := uuid.New().String()
	jwt, err := auth.GenerateJWT(user.ID, uid, res)
	if err != nil {
		logger.LOG.Error("生成JWT失败", "error", err)
		return nil, err
	}
	_ = u.cacheLocal.Set(uid, jwt, 7300)
	res.Token = uid
	res.Power = nil

	// 删除已使用的挑战
	_ = u.cacheLocal.Delete(challenge)

	return models.NewJsonResponse(200, "登录成功", res), nil
}

// Register 用户注册
func (u *UserService) Register(req *request.UserRegisterRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 检查系统是否允许注册（第一个用户注册除外，用于系统初始化）
	userCount, err := u.factory.User().Count(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询用户数量失败", "error", err)
		return nil, fmt.Errorf("系统错误")
	}

	// 如果不是第一个用户，需要检查注册配置
	if userCount > 0 {
		allowRegister, err := u.factory.SysConfig().GetByKey(ctx, "allow_register")
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LOG.Error("查询注册配置失败", "error", err)
			return nil, fmt.Errorf("系统错误")
		}
		if allowRegister == nil || allowRegister.Value != "true" {
			return nil, fmt.Errorf("系统已关闭用户注册功能，请联系管理员")
		}
	}

	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}
	decrypt, err := util.Decrypt(challengeId, req.Password)
	if err != nil {
		logger.LOG.Error("密码挑战验证失败", "error", err)
		return nil, err
	}
	psw := string(decrypt)
	if req.Username == "" || psw == "" {
		return nil, fmt.Errorf("用户名或密码不能为空")
	}
	user, err := u.factory.User().GetByUserName(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询用户失败", "error", err)
		return nil, err
	}
	if user != nil {
		return nil, fmt.Errorf("用户已存在")
	}
	v7, err := uuid.NewV7()
	if err != nil {
		logger.LOG.Error("生成UUID失败", "error", err)
		return nil, err
	}
	password, err := util.GeneratePassword(psw)
	if err != nil {
		logger.LOG.Error("生成密码失败", "error", err)
		return nil, err
	}
	isFirstUse := userCount == 0

	var groupID int
	if isFirstUse {
		groupID = 1
	} else {
		group, err := u.factory.Group().GetDefaultGroup(ctx)
		if err != nil {
			logger.LOG.Error("查询默认分组失败", "error", err)
			return nil, err
		}
		groupID = group.ID
		if groupID == 1 {
			logger.LOG.Error("默认组不能是管理员组", "group_id", groupID)
			return nil, fmt.Errorf("系统配置错误：默认组不能是管理员组，请联系管理员")
		}
	}

	group, err := u.factory.Group().GetByID(ctx, groupID)
	if err != nil {
		if isFirstUse && errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LOG.Info("首次使用，自动创建管理员组", "group_id", groupID)
			adminGroup := &models.Group{
				ID:           1,
				Name:         "管理员",
				GroupDefault: 0,
				Space:        0,
				CreatedAt:    custom_type.Now(),
			}
			if err = u.factory.Group().Create(ctx, adminGroup); err != nil {
				logger.LOG.Error("创建管理员组失败", "error", err)
				return nil, fmt.Errorf("创建管理员组失败: %w", err)
			}
			group = adminGroup

			if err = u.initDefaultPowersForAdminGroup(ctx); err != nil {
				logger.LOG.Warn("初始化默认权限失败", "error", err)
			}
			if err = u.InitEnterprisePowers(ctx); err != nil {
				logger.LOG.Warn("初始化企业权限失败", "error", err)
			}
		} else {
			logger.LOG.Error("查询组信息失败", "error", err)
			return nil, err
		}
	}

	var userSpace int64
	var userSpaceUnlimited bool
	if isFirstUse {
		userSpaceUnlimited = true
		userSpace = 0
	} else {
		userSpace = group.Space
		if userSpace == 0 {
			userSpaceUnlimited = true
		}
	}

	user = &models.UserInfo{
		ID:             v7.String(),
		Name:           req.Nickname,
		UserName:       req.Username,
		Password:       password,
		Email:          req.Email,
		Phone:          req.Phone,
		GroupID:        groupID,
		CreatedAt:      custom_type.Now(),
		Space:          userSpace,
		SpaceUnlimited: userSpaceUnlimited,
		FilePassword:   "",
		FreeSpace:      userSpace,
		State:          0,
	}
	err = u.factory.User().Create(ctx, user)
	if err != nil {
		logger.LOG.Error("创建用户失败", "error", err)
		return nil, err
	}
	virtualPath := &models.VirtualPath{
		UserID:      user.ID,
		Path:        "home",
		ParentLevel: "",
		CreatedTime: custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}
	err = u.factory.VirtualPath().Create(ctx, virtualPath)
	if err != nil {
		logger.LOG.Error("创建目录失败", "error", err)
		return nil, err
	}
	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "注册成功", user), nil
}

// initDefaultPowersForAdminGroup 初始化默认权限并分配给管理员组
func (u *UserService) initDefaultPowersForAdminGroup(ctx context.Context) error {
	logger.LOG.Info("首次注册：开始初始化所有权限并分配给管理员组")

	allPowers := []struct {
		Name           string
		Description    string
		Characteristic string
	}{
		{"用户查看", "查看系统所有用户", "user:get"},
		{"用户修改", "修改系统用户信息", "user:update"},
		{"用户删除", "删除系统用户", "user:delete"},
		{"用户停用", "暂停用户所有功能", "user:state"},
		{"用户空间分配", "分配用户可用空间大小", "user:space"},
		{"修改其他用户信息", "修改其他用户信息，包括密码", "user:update:else"},
		{"用户密码修改", "修改用户自身密码", "user:update:password"},
		{"挂载磁盘", "挂载系统可用磁盘", "disk:mount"},
		{"删除挂载磁盘", "删除已经挂载的磁盘", "disk:delete"},
		{"查看挂载磁盘", "查看已经挂载磁盘的信息", "disk:get"},
		{"文件上传", "上传文件到磁盘", "file:upload"},
		{"重命名文件", "重命名磁盘文件", "file:rechristen"},
		{"分享文件", "创建文件分享链接", "file:share"},
		{"文件下载", "下载磁盘中的文件", "file:download"},
		{"离线下载", "离线下载文件到磁盘", "file:offLine"},
		{"文件保险箱", "加密文件的上传修改下载", "file:insurance"},
		{"文件预览", "查看文件和预览支持格式的文件", "file:preview"},
		{"用户文件密码", "设置，修改文件密码", "file:update:filePassword"},
		{"移动文件", "移动文件至其他虚拟目录", "file:move"},
		{"删除文件", "删除文件（移动到回收站）", "file:delete"},
		{"创建目录", "创建文件目录", "dir:create"},
		{"删除目录", "删除已经存在的目录", "dir:delete"},
		{"创建apikey", "创建当前用户权限的apikey", "apikey:create"},
		{"删除apikey", "删除当前用户已存在的apikey", "apikey:delete"},
		{"WebDAV访问", "允许通过WebDAV协议访问文件系统", "webdav:access"},
	}

	existingPowers, err := u.factory.Power().List(ctx, 0, 1000)
	if err != nil {
		logger.LOG.Error("查询现有权限失败", "error", err)
		return fmt.Errorf("查询现有权限失败: %w", err)
	}

	powerMap := make(map[string]*models.Power)
	for _, p := range existingPowers {
		powerMap[p.Characteristic] = p
	}

	powerIDs := make([]int, 0, len(allPowers))
	maxID := 0
	for _, p := range existingPowers {
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	for _, dp := range allPowers {
		var power *models.Power

		if existingPower, ok := powerMap[dp.Characteristic]; ok {
			power = existingPower
			logger.LOG.Debug("权限已存在，跳过创建", "characteristic", dp.Characteristic)
		} else {
			maxID++
			power = &models.Power{
				ID:             maxID,
				Name:           dp.Name,
				Description:    dp.Description,
				Characteristic: dp.Characteristic,
				CreatedAt:      custom_type.Now(),
			}

			if err = u.factory.Power().Create(ctx, power); err != nil {
				logger.LOG.Error("创建权限失败", "error", err, "characteristic", dp.Characteristic)
				return fmt.Errorf("创建权限失败: %w", err)
			}

			logger.LOG.Info("创建默认权限", "name", dp.Name, "characteristic", dp.Characteristic, "id", maxID)
			powerMap[dp.Characteristic] = power
		}

		powerIDs = append(powerIDs, power.ID)
	}

	groupPowers := make([]*models.GroupPower, 0, len(powerIDs))
	for _, powerID := range powerIDs {
		groupPowers = append(groupPowers, &models.GroupPower{
			GroupID: 1,
			PowerID: powerID,
		})
	}

	if len(groupPowers) > 0 {
		if err = u.factory.GroupPower().BatchCreate(ctx, groupPowers); err != nil {
			logger.LOG.Error("分配权限给管理员组失败", "error", err)
			return fmt.Errorf("分配权限给管理员组失败: %w", err)
		}
		logger.LOG.Info("成功将默认权限分配给管理员组", "count", len(groupPowers))
	}

	return nil
}

// InitEnterprisePowers 初始化企业相关权限
// 创建 enterprise:* 命名空间的权限记录（幂等，已存在则跳过）
// 同时确保所有管理员角色拥有所有企业权限（补齐缺失的权限）
func (u *UserService) InitEnterprisePowers(ctx context.Context) error {
	logger.LOG.Info("开始初始化企业权限")

	allPowers := []struct {
		Name           string
		Description    string
		Characteristic string
	}{
		{"企业管理", "管理企业设置和信息", "enterprise:manage"},
		{"邀请成员", "邀请新成员加入企业", "enterprise:member:invite"},
		{"移除成员", "从企业中移除成员", "enterprise:member:remove"},
		{"角色管理", "创建、编辑、删除企业角色", "enterprise:role:manage"},
		{"上传到共享空间", "上传文件到企业共享空间", "enterprise:space:upload"},
		{"从共享空间下载", "下载企业共享空间中的文件", "enterprise:space:download"},
		{"删除共享空间文件", "删除企业共享空间中的文件", "enterprise:space:delete"},
		{"查看审计日志", "查看企业审计日志", "enterprise:audit:view"},
		{"查看成员列表", "查看企业成员列表", "enterprise:member:view"},
	}

	existingPowers, err := u.factory.Power().List(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("查询现有权限失败: %w", err)
	}

	powerMap := make(map[string]*models.Power)
	maxID := 0
	for _, p := range existingPowers {
		powerMap[p.Characteristic] = p
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	created := 0
	for _, dp := range allPowers {
		if _, ok := powerMap[dp.Characteristic]; ok {
			continue
		}
		maxID++
		power := &models.Power{
			ID:             maxID,
			Name:           dp.Name,
			Description:    dp.Description,
			Characteristic: dp.Characteristic,
			CreatedAt:      custom_type.Now(),
		}
		if err = u.factory.Power().Create(ctx, power); err != nil {
			logger.LOG.Error("创建企业权限失败", "error", err, "characteristic", dp.Characteristic)
			return fmt.Errorf("创建企业权限失败: %w", err)
		}
		powerMap[dp.Characteristic] = power
		created++
		logger.LOG.Info("创建企业权限", "name", dp.Name, "characteristic", dp.Characteristic, "id", maxID)
	}

	// 确保所有管理员角色拥有所有企业权限（补齐缺失的权限）
	var adminRoles []*models.EnterpriseRole
	if err := u.factory.DB().Where("is_admin = 1").Find(&adminRoles).Error; err != nil {
		logger.LOG.Error("查询管理员角色失败", "error", err)
	} else {
		for _, role := range adminRoles {
			rolePowers, err := u.factory.EnterpriseRolePower().GetByRoleID(ctx, role.ID)
			if err != nil {
				logger.LOG.Error("查询角色权限失败", "error", err, "roleID", role.ID)
				continue
			}
			hasPower := make(map[int]bool)
			for _, rp := range rolePowers {
				hasPower[rp.PowerID] = true
			}
			var missing []*models.EnterpriseRolePower
			for _, dp := range allPowers {
				if p, ok := powerMap[dp.Characteristic]; ok {
					if !hasPower[p.ID] {
						missing = append(missing, &models.EnterpriseRolePower{
							RoleID:  role.ID,
							PowerID: p.ID,
						})
					}
				}
			}
			if len(missing) > 0 {
				if err := u.factory.EnterpriseRolePower().BatchCreate(ctx, missing); err != nil {
					logger.LOG.Error("补齐角色权限失败", "error", err, "roleID", role.ID)
				} else {
					logger.LOG.Info("已为管理员角色补齐权限", "roleID", role.ID, "count", len(missing))
				}
			}
		}
	}

	if created == 0 {
		logger.LOG.Info("企业权限已全部存在，无需创建")
	} else {
		logger.LOG.Info("企业权限初始化完成", "created", created)
	}
	return nil
}

// Challenge 密码挑战
func (u *UserService) Challenge() (*models.JsonResponse, error) {
	pair, err := util.GenerateKeyPair()
	if err != nil {
		logger.LOG.Error("生成密钥对失败", "error", err)
		return nil, err
	}
	uid := uuid.NewString()
	err = u.cacheLocal.Set(uid, pair.PrivateKey, 60)
	if err != nil {
		logger.LOG.Error("缓存密钥对失败", "error", err)
		return nil, err
	}
	m := map[string]string{
		"publicKey": pair.PublicKey,
		"id":        uid,
	}
	return models.NewJsonResponse(200, "ok", m), nil
}

// SysInit 查询系统是否初次使用和注册配置
func (u *UserService) SysInit() (*models.JsonResponse, error) {
	ctx := context.Background()
	count, err := u.factory.User().Count(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询用户数量失败", "error", err)
		return nil, err
	}
	isFirstUse := count == 0

	allowRegister := true
	if !isFirstUse {
		allowRegisterConfig, err := u.factory.SysConfig().GetByKey(ctx, "allow_register")
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			allowRegister = false
		} else if allowRegisterConfig != nil {
			allowRegister = allowRegisterConfig.Value == "true"
		} else {
			allowRegister = false
		}
	}

	result := map[string]interface{}{
		"is_first_use":   isFirstUse,
		"allow_register": allowRegister,
	}
	return models.NewJsonResponse(200, "ok", result), nil
}

// UpdateUser 修改用户信息
func (u *UserService) UpdateUser(req *request.UserUpdateRequest) (*models.JsonResponse, error) {
	ctx := context.Background()
	user, err := u.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		user.UserName = req.Username
	}
	if req.Nickname != "" {
		user.Name = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if err := u.factory.User().Update(ctx, user); err != nil {
		return nil, err
	}
	return models.NewJsonResponse(200, "ok", nil), nil
}

// UpdatePassword 修改用户密码
func (u *UserService) UpdatePassword(req *request.UserUpdatePasswordRequest) (*models.JsonResponse, error) {
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	decryptOld, err := util.Decrypt(challengeId, req.OldPasswd)
	if err != nil {
		logger.LOG.Error("旧密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	oldPsw := string(decryptOld)

	decryptNew, err := util.Decrypt(challengeId, req.NewPasswd)
	if err != nil {
		logger.LOG.Error("新密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	newPsw := string(decryptNew)

	ctx := context.Background()
	user, err := u.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if !util.CheckPassword(user.Password, oldPsw) {
		return nil, fmt.Errorf("密码错误")
	}
	password, err := util.GeneratePassword(newPsw)
	if err != nil {
		return nil, err
	}
	user.Password = password
	if err := u.factory.User().Update(ctx, user); err != nil {
		return nil, err
	}

	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}

// SetFilePassword 设置文件密码
func (u *UserService) SetFilePassword(req *request.UserSetFilePasswordRequest) (*models.JsonResponse, error) {
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	decrypt, err := util.Decrypt(challengeId, req.Passwd)
	if err != nil {
		logger.LOG.Error("密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	psw := string(decrypt)

	ctx := context.Background()
	user, err := u.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if psw == "" {
		return nil, fmt.Errorf("密码不能为空")
	}
	user.FilePassword, err = util.GeneratePassword(psw)
	if err != nil {
		return nil, err
	}
	if err := u.factory.User().Update(ctx, user); err != nil {
		return nil, err
	}

	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}

// UpdateFilePassword 修改文件密码
func (u *UserService) UpdateFilePassword(req *request.UserUpdatePasswordRequest) (*models.JsonResponse, error) {
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	decryptOld, err := util.Decrypt(challengeId, req.OldPasswd)
	if err != nil {
		logger.LOG.Error("旧密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	oldPsw := string(decryptOld)

	decryptNew, err := util.Decrypt(challengeId, req.NewPasswd)
	if err != nil {
		logger.LOG.Error("新密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	newPsw := string(decryptNew)

	ctx := context.Background()
	user, err := u.factory.User().GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if !util.CheckPassword(user.FilePassword, oldPsw) {
		return nil, fmt.Errorf("密码错误")
	}
	password, err := util.GeneratePassword(newPsw)
	if err != nil {
		return nil, err
	}
	user.FilePassword = password
	if err := u.factory.User().Update(ctx, user); err != nil {
		return nil, err
	}

	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}

// GenerateApiKey 生成API Key
func (u *UserService) GenerateApiKey(req *request.GenerateApiKeyRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	apiKeyStr := uuid.Must(uuid.NewV7()).String()

	keyPair, err := util.GenerateKeyPair()
	if err != nil {
		logger.LOG.Error("生成密钥对失败", "error", err)
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	s3SecretKey := uuid.Must(uuid.NewV7()).String() + uuid.Must(uuid.NewV7()).String()

	var expiresAt custom_type.JsonTime
	if req.ExpiresDays > 0 {
		now := custom_type.Now().ToTime()
		expiresAt = custom_type.JsonTime(now.AddDate(0, 0, req.ExpiresDays))
	} else {
		expiresAt = custom_type.JsonTime{}
	}

	apiKey := &models.ApiKey{
		UserID:      userID,
		Key:         apiKeyStr,
		PrivateKey:  keyPair.PrivateKey,
		S3SecretKey: s3SecretKey,
		ExpiresAt:   expiresAt,
		CreatedAt:   custom_type.Now(),
	}

	if err := u.factory.ApiKey().Create(ctx, apiKey); err != nil {
		logger.LOG.Error("保存API Key失败", "error", err)
		return nil, fmt.Errorf("保存API Key失败: %w", err)
	}

	logger.LOG.Info("API Key已生成", "userID", userID, "apiKeyID", apiKey.ID)

	var expiresAtResp interface{} = nil
	if !expiresAt.IsZero() {
		expiresAtResp = expiresAt
	}

	return models.NewJsonResponse(200, "API Key生成成功", map[string]interface{}{
		"id":            apiKey.ID,
		"key":           apiKeyStr,
		"public_key":    keyPair.PublicKey,
		"s3_secret_key": s3SecretKey,
		"expires_at":    expiresAtResp,
		"created_at":    apiKey.CreatedAt,
	}), nil
}

// ListApiKeys 获取用户的API Key列表
func (u *UserService) ListApiKeys(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	apiKeys, err := u.factory.ApiKey().List(ctx, userID, 0, 100)
	if err != nil {
		logger.LOG.Error("查询API Key列表失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("查询API Key列表失败: %w", err)
	}

	items := make([]map[string]interface{}, 0, len(apiKeys))
	for _, key := range apiKeys {
		maskedKey := maskApiKey(key.Key)

		var expiresAt interface{} = nil
		if !key.ExpiresAt.IsZero() {
			expiresAt = key.ExpiresAt
		}

		item := map[string]interface{}{
			"id":         key.ID,
			"key":        maskedKey,
			"expires_at": expiresAt,
			"created_at": key.CreatedAt,
			"is_expired": false,
		}

		if !key.ExpiresAt.IsZero() {
			expiresTime := time.Time(key.ExpiresAt)
			if expiresTime.Before(time.Now()) {
				item["is_expired"] = true
			}
		}

		items = append(items, item)
	}

	return models.NewJsonResponse(200, "获取成功", items), nil
}

// DeleteApiKey 删除API Key
func (u *UserService) DeleteApiKey(req *request.DeleteApiKeyRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	apiKey, err := u.factory.ApiKey().GetByID(ctx, req.ApiKeyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.NewJsonResponse(404, "API Key不存在", nil), nil
		}
		logger.LOG.Error("获取API Key失败", "error", err, "apiKeyID", req.ApiKeyID)
		return nil, fmt.Errorf("获取API Key失败: %w", err)
	}

	if apiKey.UserID != userID {
		logger.LOG.Warn("用户尝试删除他人的API Key", "userID", userID, "apiKeyID", req.ApiKeyID)
		return models.NewJsonResponse(403, "无权操作此API Key", nil), nil
	}

	if err := u.factory.ApiKey().Delete(ctx, req.ApiKeyID); err != nil {
		logger.LOG.Error("删除API Key失败", "error", err, "apiKeyID", req.ApiKeyID)
		return nil, fmt.Errorf("删除API Key失败: %w", err)
	}

	logger.LOG.Info("API Key已删除", "userID", userID, "apiKeyID", req.ApiKeyID)
	return models.NewJsonResponse(200, "API Key已删除", nil), nil
}

func maskApiKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}

func (u *UserService) GetUserInfo(userID string) (*models.JsonResponse, error) {
	id, err := u.factory.User().GetByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	return models.NewJsonResponse(200, "ok", response.UserInfoResponse{
		ID:             id.ID,
		Name:           id.Name,
		Email:          id.Email,
		Phone:          id.Phone,
		GroupID:        id.GroupID,
		State:          id.State,
		Space:          id.Space,
		FreeSpace:      id.FreeSpace,
		SpaceUnlimited: id.SpaceUnlimited,
		UserName:       id.UserName,
	}), nil
}
