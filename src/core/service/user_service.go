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
	ctx := context.Background()
	// 验证用户名和密码
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
	group, err := u.factory.Group().GetDefaultGroup(ctx)
	if err != nil {
		logger.LOG.Error("查询默认分组失败", "error", err)
		return nil, err
	}
	user = &models.UserInfo{
		ID:           v7.String(),
		Name:         req.Nickname,
		UserName:     req.Username,
		Password:     password,
		Email:        req.Email,
		Phone:        req.Phone,
		GroupID:      group.ID,
		CreatedAt:    custom_type.Now(),
		Space:        0,
		FilePassword: "",
		FreeSpace:    0,
		State:        0,
	}
	init, err := u.SysInit()
	if err != nil {
		logger.LOG.Error("系统初始化失败", "error", err)
		return nil, err
	}
	b := init.Data.(bool)
	if !b {
		user.GroupID = 1
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
	// 删除已使用的挑战
	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "注册成功", user), nil
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

// SysInit 查询系统是否初次使用
func (u *UserService) SysInit() (*models.JsonResponse, error) {
	count, err := u.factory.User().Count(context.Background())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询用户数量失败", "error", err)
		return nil, err
	}
	return models.NewJsonResponse(200, "ok", count == 0), nil
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
	// 验证挑战是否有效
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	// 解密旧密码
	decryptOld, err := util.Decrypt(challengeId, req.OldPasswd)
	if err != nil {
		logger.LOG.Error("旧密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	oldPsw := string(decryptOld)

	// 解密新密码
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

	// 删除已使用的挑战
	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}

// SetFilePassword 设置文件密码
func (u *UserService) SetFilePassword(req *request.UserSetFilePasswordRequest) (*models.JsonResponse, error) {
	// 验证挑战是否有效
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	// 解密密码
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

	// 删除已使用的挑战
	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}

// UpdateFilePassword 修改文件密码
func (u *UserService) UpdateFilePassword(req *request.UserUpdatePasswordRequest) (*models.JsonResponse, error) {
	// 验证挑战是否有效
	get, err := u.cacheLocal.Get(req.Challenge)
	if err != nil {
		logger.LOG.Error("获取缓存失败", "error", err)
		return nil, fmt.Errorf("验证已过期")
	}
	challengeId := get.(string)
	if challengeId == "" {
		return nil, fmt.Errorf("验证已过期")
	}

	// 解密旧密码
	decryptOld, err := util.Decrypt(challengeId, req.OldPasswd)
	if err != nil {
		logger.LOG.Error("旧密码解密失败", "error", err)
		return nil, fmt.Errorf("密码验证失败")
	}
	oldPsw := string(decryptOld)

	// 解密新密码
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

	// 删除已使用的挑战
	_ = u.cacheLocal.Delete(req.Challenge)

	return models.NewJsonResponse(200, "ok", nil), nil
}
