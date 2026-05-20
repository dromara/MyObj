package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"math/big"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnterpriseService struct {
	factory *impl.RepositoryFactory
}

func NewEnterpriseService(factory *impl.RepositoryFactory) *EnterpriseService {
	return &EnterpriseService{factory: factory}
}

func (s *EnterpriseService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// CreateEnterprise 创建企业
func (s *EnterpriseService) CreateEnterprise(req *request.CreateEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	inviteCode, err := generateInviteCode()
	if err != nil {
		return models.NewJsonResponse(500, "生成邀请码失败", nil), err
	}

	// 获取创建者信息，判断是否为超级管理员
	creator, err := s.factory.User().GetByID(ctx, userID)
	if err != nil {
		return models.NewJsonResponse(500, "获取用户信息失败", nil), err
	}

	// 根据创建者身份设置企业空间
	// 超级管理员(GroupID=1)创建的企业：使用全局默认企业空间配置
	// 非超级管理员创建的企业：空间为0（需要超管手动分配）
	var enterpriseSpace int64
	var spaceUnlimited bool
	if creator.GroupID == 1 {
		// 超级管理员创建，读取全局配置
		cfg, _ := s.factory.SysConfig().GetByKey(ctx, "default_enterprise_space")
		if cfg != nil {
			fmt.Sscanf(cfg.Value, "%d", &enterpriseSpace)
		}
		// 如果配置为0，表示无限空间
		if enterpriseSpace == 0 {
			spaceUnlimited = true
		}
	}
	// 非超管创建：enterpriseSpace = 0, spaceUnlimited = false（即无空间）

	enterpriseID := uuid.New().String()
	now := custom_type.Now()

	var result *models.JsonResponse
	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)

		enterprise := &models.Enterprise{
			ID:             enterpriseID,
			Name:           req.Name,
			Logo:           req.Logo,
			Description:    req.Description,
			CreatorID:      userID,
			Space:          enterpriseSpace,
			FreeSpace:      enterpriseSpace,
			SpaceUnlimited: spaceUnlimited,
			InviteCode:     inviteCode,
			State:          0,
			CreatedAt:      now,
		}
		if err := txFactory.Enterprise().Create(ctx, enterprise); err != nil {
			return err
		}

		adminRole := &models.EnterpriseRole{
			ID:           uuid.New().String(),
			EnterpriseID: enterpriseID,
			Name:         "管理员",
			IsDefault:    0,
			IsAdmin:      1,
			CreatedAt:    now,
		}
		if err := txFactory.EnterpriseRole().Create(ctx, adminRole); err != nil {
			return err
		}

		allPowers, _ := txFactory.Power().List(ctx, 0, 1000)
		var adminRolePowers []*models.EnterpriseRolePower
		for _, p := range allPowers {
			if len(p.Characteristic) >= 11 && p.Characteristic[:11] == "enterprise:" {
				adminRolePowers = append(adminRolePowers, &models.EnterpriseRolePower{
					RoleID:  adminRole.ID,
					PowerID: p.ID,
				})
			}
		}
		if len(adminRolePowers) > 0 {
			if err := txFactory.EnterpriseRolePower().BatchCreate(ctx, adminRolePowers); err != nil {
				return err
			}
		}

		defaultRole := &models.EnterpriseRole{
			ID:           uuid.New().String(),
			EnterpriseID: enterpriseID,
			Name:         "普通成员",
			IsDefault:    1,
			IsAdmin:      0,
			CreatedAt:    now,
		}
		if err := txFactory.EnterpriseRole().Create(ctx, defaultRole); err != nil {
			return err
		}

		member := &models.EnterpriseMember{
			ID:           uuid.New().String(),
			EnterpriseID: enterpriseID,
			UserID:       userID,
			RoleID:       adminRole.ID,
			JoinedAt:     now,
			Status:       0,
		}
		if err := txFactory.EnterpriseMember().Create(ctx, member); err != nil {
			return err
		}

		result = models.NewJsonResponse(200, "创建企业成功", map[string]interface{}{
			"enterprise_id": enterpriseID,
			"invite_code":   inviteCode,
		})
		return nil
	})

	if err != nil {
		return models.NewJsonResponse(500, "创建企业失败", nil), err
	}
	return result, nil
}

// GetMyEnterprises 获取我加入的企业列表
// getUserPowers 获取用户在企业中的权限列表
func (s *EnterpriseService) getUserPowers(ctx context.Context, member *models.EnterpriseMember) []string {
	if member == nil {
		return nil
	}
	role, err := s.factory.EnterpriseRole().GetByID(ctx, member.RoleID)
	if err != nil || role == nil {
		return nil
	}
	rolePowers, err := s.factory.EnterpriseRolePower().GetByRoleID(ctx, role.ID)
	if err != nil {
		return nil
	}
	var powers []string
	for _, rp := range rolePowers {
		power, err := s.factory.Power().GetByID(ctx, rp.PowerID)
		if err == nil && power != nil {
			powers = append(powers, power.Characteristic)
		}
	}
	return powers
}

func (s *EnterpriseService) GetMyEnterprises(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	members, err := s.factory.EnterpriseMember().ListByUserID(ctx, userID)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}

	var list []*response.EnterpriseResponse
	for _, m := range members {
		enterprise, err := s.factory.Enterprise().GetByID(ctx, m.EnterpriseID)
		if err != nil {
			continue
		}
		memberCount, _ := s.factory.EnterpriseMember().CountByEnterpriseID(ctx, m.EnterpriseID)
		role, _ := s.factory.EnterpriseRole().GetByID(ctx, m.RoleID)
		roleName := ""
		isAdmin := 0
		if role != nil {
			roleName = role.Name
			isAdmin = role.IsAdmin
		}
		powers := s.getUserPowers(ctx, m)
		list = append(list, &response.EnterpriseResponse{
			ID:             enterprise.ID,
			Name:           enterprise.Name,
			Logo:           enterprise.Logo,
			Description:    enterprise.Description,
			CreatorID:      enterprise.CreatorID,
			Space:          enterprise.Space,
			FreeSpace:      enterprise.FreeSpace,
			SpaceUnlimited: enterprise.SpaceUnlimited,
			InviteCode:     enterprise.InviteCode,
			State:          enterprise.State,
			CreatedAt:      enterprise.CreatedAt,
			MemberCount:    memberCount,
			Role:           roleName,
			IsAdmin:        isAdmin,
			Powers:         powers,
		})
	}

	return models.NewJsonResponse(200, "ok", list), nil
}

// GetEnterpriseInfo 获取企业详情
func (s *EnterpriseService) GetEnterpriseInfo(enterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	enterprise, err := s.factory.Enterprise().GetByID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	memberCount, _ := s.factory.EnterpriseMember().CountByEnterpriseID(ctx, enterpriseID)
	member, _ := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	roleName := ""
	isAdmin := 0
	if member != nil {
		role, _ := s.factory.EnterpriseRole().GetByID(ctx, member.RoleID)
		if role != nil {
			roleName = role.Name
			isAdmin = role.IsAdmin
		}
	}
	powers := s.getUserPowers(ctx, member)

	return models.NewJsonResponse(200, "ok", &response.EnterpriseResponse{
		ID:             enterprise.ID,
		Name:           enterprise.Name,
		Logo:           enterprise.Logo,
		Description:    enterprise.Description,
		CreatorID:      enterprise.CreatorID,
		Space:          enterprise.Space,
		FreeSpace:      enterprise.FreeSpace,
		SpaceUnlimited: enterprise.SpaceUnlimited,
		InviteCode:     enterprise.InviteCode,
		State:          enterprise.State,
		CreatedAt:      enterprise.CreatedAt,
		MemberCount:    memberCount,
		Role:           roleName,
		IsAdmin:        isAdmin,
		Powers:         powers,
	}), nil
}

// UpdateEnterprise 更新企业信息
func (s *EnterpriseService) UpdateEnterprise(req *request.UpdateEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证是否为企业管理员
	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	if req.Name != "" {
		enterprise.Name = req.Name
	}
	enterprise.Description = req.Description
	enterprise.Logo = req.Logo

	if err := s.factory.Enterprise().Update(ctx, enterprise); err != nil {
		return models.NewJsonResponse(500, "更新失败", nil), err
	}

	return models.NewJsonResponse(200, "更新成功", nil), nil
}

// SwitchEnterprise 切换当前企业/个人空间
func (s *EnterpriseService) SwitchEnterprise(req *request.SwitchEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	user, err := s.factory.User().GetByID(ctx, userID)
	if err != nil {
		return models.NewJsonResponse(500, "查询用户失败", nil), err
	}

	// 如果切换到企业，验证用户是否为该企业成员
	if req.EnterpriseID != "" {
		member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
		if err != nil || member == nil {
			return models.NewJsonResponse(403, "您不是该企业成员", nil), fmt.Errorf("not a member")
		}
	}

	user.CurrentEnterpriseID = req.EnterpriseID
	if err := s.factory.User().Update(ctx, user); err != nil {
		return models.NewJsonResponse(500, "切换失败", nil), err
	}

	spaceType := "个人空间"
	if req.EnterpriseID != "" {
		enterprise, _ := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
		if enterprise != nil {
			spaceType = enterprise.Name
		}
	}

	return models.NewJsonResponse(200, "切换成功", map[string]interface{}{
		"current_enterprise_id": req.EnterpriseID,
		"space_type":            spaceType,
	}), nil
}

// InviteMember 直接邀请成员
func (s *EnterpriseService) InviteMember(req *request.InviteMemberRequest, inviterID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, inviterID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	// 查找目标用户
	targetUser, err := s.factory.User().GetByUserName(ctx, req.UserName)
	if err != nil {
		return models.NewJsonResponse(404, "用户不存在", nil), err
	}

	// 检查是否已经是成员
	existing, _ := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, targetUser.ID)
	if existing != nil {
		return models.NewJsonResponse(400, "该用户已是企业成员", nil), nil
	}

	invite := &models.EnterpriseInvite{
		ID:           uuid.New().String(),
		EnterpriseID: req.EnterpriseID,
		InviterID:    inviterID,
		InviteeID:    targetUser.ID,
		Type:         1,
		Status:       0,
		ExpireAt:     expireTime(7), // 7天过期
		CreatedAt:    custom_type.Now(),
	}

	if err := s.factory.EnterpriseInvite().Create(ctx, invite); err != nil {
		return models.NewJsonResponse(500, "创建邀请失败", nil), err
	}

	return models.NewJsonResponse(200, "邀请已发送", nil), nil
}

// GetInviteCode 获取企业邀请码
func (s *EnterpriseService) GetInviteCode(enterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, enterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"invite_code": enterprise.InviteCode,
		"invite_link": enterprise.InviteLink,
	}), nil
}

// RefreshInviteCode 刷新企业邀请码
func (s *EnterpriseService) RefreshInviteCode(enterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, enterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	newCode, err := generateInviteCode()
	if err != nil {
		return models.NewJsonResponse(500, "生成邀请码失败", nil), err
	}

	enterprise.InviteCode = newCode
	if err := s.factory.Enterprise().Update(ctx, enterprise); err != nil {
		return models.NewJsonResponse(500, "刷新失败", nil), err
	}

	return models.NewJsonResponse(200, "刷新成功", map[string]interface{}{
		"invite_code": newCode,
	}), nil
}

// JoinEnterprise 通过邀请码加入企业
func (s *EnterpriseService) JoinEnterprise(req *request.JoinEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	enterprise, err := s.factory.Enterprise().GetByInviteCode(ctx, req.InviteCode)
	if err != nil {
		return models.NewJsonResponse(404, "邀请码无效", nil), err
	}

	if enterprise.State != 0 {
		return models.NewJsonResponse(400, "企业已禁用", nil), nil
	}

	// 检查是否已经是成员
	existing, _ := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterprise.ID, userID)
	if existing != nil {
		return models.NewJsonResponse(400, "您已是该企业成员", nil), nil
	}

	// 获取默认角色
	defaultRole, err := s.factory.EnterpriseRole().GetDefaultByEnterpriseID(ctx, enterprise.ID)
	if err != nil {
		return models.NewJsonResponse(500, "获取默认角色失败", nil), err
	}

	member := &models.EnterpriseMember{
		ID:           uuid.New().String(),
		EnterpriseID: enterprise.ID,
		UserID:       userID,
		RoleID:       defaultRole.ID,
		JoinedAt:     custom_type.Now(),
		Status:       0,
	}

	if err := s.factory.EnterpriseMember().Create(ctx, member); err != nil {
		return models.NewJsonResponse(500, "加入失败", nil), err
	}

	return models.NewJsonResponse(200, "加入成功", map[string]interface{}{
		"enterprise_id":   enterprise.ID,
		"enterprise_name": enterprise.Name,
	}), nil
}

// GetPendingInvites 获取待处理邀请列表
func (s *EnterpriseService) GetPendingInvites(userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	invites, err := s.factory.EnterpriseInvite().ListPendingByInviteeID(ctx, userID)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}

	var list []*response.EnterpriseInviteResponse
	for _, inv := range invites {
		enterprise, _ := s.factory.Enterprise().GetByID(ctx, inv.EnterpriseID)
		inviter, _ := s.factory.User().GetByID(ctx, inv.InviterID)
		enterpriseName := ""
		inviterName := ""
		if enterprise != nil {
			enterpriseName = enterprise.Name
		}
		if inviter != nil {
			inviterName = inviter.Name
		}
		list = append(list, &response.EnterpriseInviteResponse{
			ID:             inv.ID,
			EnterpriseID:   inv.EnterpriseID,
			EnterpriseName: enterpriseName,
			InviterID:      inv.InviterID,
			InviterName:    inviterName,
			Status:         inv.Status,
			CreatedAt:      inv.CreatedAt,
		})
	}

	return models.NewJsonResponse(200, "ok", list), nil
}

// AcceptInvite 接受邀请
func (s *EnterpriseService) AcceptInvite(inviteID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	invite, err := s.factory.EnterpriseInvite().GetByID(ctx, inviteID)
	if err != nil {
		return models.NewJsonResponse(404, "邀请不存在", nil), err
	}

	if invite.Status != 0 {
		return models.NewJsonResponse(400, "邀请已处理", nil), nil
	}

	if invite.InviteeID != userID {
		return models.NewJsonResponse(403, "无权操作", nil), nil
	}

	// 检查企业状态
	enterprise, _ := s.factory.Enterprise().GetByID(ctx, invite.EnterpriseID)
	if enterprise != nil && enterprise.State != 0 {
		return models.NewJsonResponse(400, "企业已禁用", nil), nil
	}

	// 检查是否已是成员
	existing, _ := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, invite.EnterpriseID, userID)
	if existing != nil {
		return models.NewJsonResponse(400, "您已是该企业成员", nil), nil
	}

	defaultRole, err := s.factory.EnterpriseRole().GetDefaultByEnterpriseID(ctx, invite.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "获取默认角色失败", nil), err
	}

	member := &models.EnterpriseMember{
		ID:           uuid.New().String(),
		EnterpriseID: invite.EnterpriseID,
		UserID:       userID,
		RoleID:       defaultRole.ID,
		JoinedAt:     custom_type.Now(),
		Status:       0,
	}
	if err := s.factory.EnterpriseMember().Create(ctx, member); err != nil {
		return models.NewJsonResponse(500, "加入失败", nil), err
	}

	invite.Status = 1
	if err := s.factory.EnterpriseInvite().Update(ctx, invite); err != nil {
		return models.NewJsonResponse(500, "更新邀请状态失败", nil), err
	}

	return models.NewJsonResponse(200, "加入成功", nil), nil
}

// GetMemberList 获取企业成员列表
func (s *EnterpriseService) GetMemberList(req *request.EnterpriseListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkMember(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "无权访问", nil), err
	}

	offset := (req.Page - 1) * req.PageSize
	members, err := s.factory.EnterpriseMember().ListByEnterpriseID(ctx, req.EnterpriseID, offset, req.PageSize)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}

	total, _ := s.factory.EnterpriseMember().CountByEnterpriseID(ctx, req.EnterpriseID)

	var list []*response.EnterpriseMemberResponse
	for _, m := range members {
		user, _ := s.factory.User().GetByID(ctx, m.UserID)
		role, _ := s.factory.EnterpriseRole().GetByID(ctx, m.RoleID)
		userName := ""
		userAvatar := ""
		if user != nil {
			userName = user.UserName
			if user.Name != "" {
				userName = user.Name
			}
		}
		roleName := ""
		isAdmin := 0
		if role != nil {
			roleName = role.Name
			isAdmin = role.IsAdmin
		}
		list = append(list, &response.EnterpriseMemberResponse{
			ID:         m.ID,
			UserID:     m.UserID,
			UserName:   userName,
			UserAvatar: userAvatar,
			Status:     m.Status,
			RoleID:     m.RoleID,
			RoleName:   roleName,
			IsAdmin:    isAdmin,
			JoinedAt:   m.JoinedAt,
		})
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"total": total,
		"list":  list,
	}), nil
}

// UpdateMemberRole 修改成员角色
func (s *EnterpriseService) UpdateMemberRole(req *request.UpdateMemberRoleRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByID(ctx, req.MemberID)
	if err != nil {
		return models.NewJsonResponse(404, "成员不存在", nil), err
	}

	member.RoleID = req.RoleID
	if err := s.factory.EnterpriseMember().Update(ctx, member); err != nil {
		return models.NewJsonResponse(500, "修改失败", nil), err
	}

	return models.NewJsonResponse(200, "修改成功", nil), nil
}

// RemoveMember 移除成员
func (s *EnterpriseService) RemoveMember(req *request.RemoveMemberRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByID(ctx, req.MemberID)
	if err != nil {
		return models.NewJsonResponse(404, "成员不存在", nil), err
	}

	// 不能移除自己
	if member.UserID == userID {
		return models.NewJsonResponse(400, "不能移除自己", nil), nil
	}

	member.Status = 2
	if err := s.factory.EnterpriseMember().Update(ctx, member); err != nil {
		return models.NewJsonResponse(500, "移除失败", nil), err
	}

	return models.NewJsonResponse(200, "移除成功", nil), nil
}

// LeaveEnterprise 退出企业
func (s *EnterpriseService) LeaveEnterprise(req *request.LeaveEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
	if err != nil || member == nil {
		return models.NewJsonResponse(404, "您不是该企业成员", nil), err
	}

	// 企业创建者不能退出
	enterprise, _ := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if enterprise != nil && enterprise.CreatorID == userID {
		return models.NewJsonResponse(400, "企业创建者不能退出", nil), nil
	}

	// 如果是管理员角色，检查是否是最后一个管理员
	role, _ := s.factory.EnterpriseRole().GetByID(ctx, member.RoleID)
	if role != nil && role.IsAdmin == 1 {
		members, _ := s.factory.EnterpriseMember().ListByEnterpriseID(ctx, req.EnterpriseID, 0, 0)
		adminCount := 0
		for _, m := range members {
			if m.Status == 0 {
				r, _ := s.factory.EnterpriseRole().GetByID(ctx, m.RoleID)
				if r != nil && r.IsAdmin == 1 {
					adminCount++
				}
			}
		}
		if adminCount <= 1 {
			return models.NewJsonResponse(400, "最后一个管理员不能退出，请先转让管理权", nil), nil
		}
	}

	member.Status = 1
	if err := s.factory.EnterpriseMember().Update(ctx, member); err != nil {
		return models.NewJsonResponse(500, "退出失败", nil), err
	}

	return models.NewJsonResponse(200, "退出成功", nil), nil
}

// GetRoleList 获取企业角色列表
func (s *EnterpriseService) GetRoleList(enterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkMember(ctx, enterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "无权访问", nil), err
	}

	roles, err := s.factory.EnterpriseRole().ListByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}

	var list []*response.EnterpriseRoleResponse
	for _, r := range roles {
		rolePowers, _ := s.factory.EnterpriseRolePower().GetByRoleID(ctx, r.ID)
		var powerIDs []int
		for _, rp := range rolePowers {
			powerIDs = append(powerIDs, rp.PowerID)
		}
		list = append(list, &response.EnterpriseRoleResponse{
			ID:        r.ID,
			Name:      r.Name,
			IsDefault: r.IsDefault,
			IsAdmin:   r.IsAdmin,
			PowerIDs:  powerIDs,
		})
	}

	return models.NewJsonResponse(200, "ok", list), nil
}

// CreateRole 创建企业角色
func (s *EnterpriseService) CreateRole(req *request.CreateEnterpriseRoleRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	role := &models.EnterpriseRole{
		ID:           uuid.New().String(),
		EnterpriseID: req.EnterpriseID,
		Name:         req.Name,
		IsDefault:    0,
		IsAdmin:      0,
		CreatedAt:    custom_type.Now(),
	}

	if err := s.factory.EnterpriseRole().Create(ctx, role); err != nil {
		return models.NewJsonResponse(500, "创建角色失败", nil), err
	}

	// 分配权限
	if len(req.PowerIDs) > 0 {
		var rolePowers []*models.EnterpriseRolePower
		for _, pid := range req.PowerIDs {
			rolePowers = append(rolePowers, &models.EnterpriseRolePower{
				RoleID:  role.ID,
				PowerID: pid,
			})
		}
		s.factory.EnterpriseRolePower().BatchCreate(ctx, rolePowers)
	}

	return models.NewJsonResponse(200, "创建成功", map[string]interface{}{
		"role_id": role.ID,
	}), nil
}

// UpdateRole 更新企业角色
func (s *EnterpriseService) UpdateRole(req *request.UpdateEnterpriseRoleRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	role, err := s.factory.EnterpriseRole().GetByID(ctx, req.RoleID)
	if err != nil {
		return models.NewJsonResponse(404, "角色不存在", nil), err
	}

	if err := s.checkAdmin(ctx, role.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	if role.IsAdmin == 1 {
		return models.NewJsonResponse(400, "不能修改管理员角色", nil), nil
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if err := s.factory.EnterpriseRole().Update(ctx, role); err != nil {
		return models.NewJsonResponse(500, "更新失败", nil), err
	}

	// 更新权限
	if req.PowerIDs != nil {
		s.factory.EnterpriseRolePower().DeleteByRoleID(ctx, role.ID)
		if len(req.PowerIDs) > 0 {
			var rolePowers []*models.EnterpriseRolePower
			for _, pid := range req.PowerIDs {
				rolePowers = append(rolePowers, &models.EnterpriseRolePower{
					RoleID:  role.ID,
					PowerID: pid,
				})
			}
			s.factory.EnterpriseRolePower().BatchCreate(ctx, rolePowers)
		}
	}

	return models.NewJsonResponse(200, "更新成功", nil), nil
}

// DeleteRole 删除企业角色
func (s *EnterpriseService) DeleteRole(req *request.DeleteEnterpriseRoleRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	role, err := s.factory.EnterpriseRole().GetByID(ctx, req.RoleID)
	if err != nil {
		return models.NewJsonResponse(404, "角色不存在", nil), err
	}

	if err := s.checkAdmin(ctx, role.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	if role.IsAdmin == 1 {
		return models.NewJsonResponse(400, "不能删除管理员角色", nil), nil
	}
	if role.IsDefault == 1 {
		return models.NewJsonResponse(400, "不能删除默认角色", nil), nil
	}

	// 检查是否有成员使用此角色
	members, _ := s.factory.EnterpriseMember().ListByEnterpriseID(ctx, role.EnterpriseID, 0, 0)
	for _, m := range members {
		if m.RoleID == role.ID && m.Status == 0 {
			return models.NewJsonResponse(400, "该角色下仍有成员，请先移除或更换成员角色", nil), nil
		}
	}

	if err := s.factory.EnterpriseRolePower().DeleteByRoleID(ctx, role.ID); err != nil {
		return models.NewJsonResponse(500, "删除角色权限失败", nil), err
	}
	if err := s.factory.EnterpriseRole().Delete(ctx, role.ID); err != nil {
		return models.NewJsonResponse(500, "删除角色失败", nil), err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// TransferOwnership 转让企业所有权
func (s *EnterpriseService) TransferOwnership(req *request.TransferOwnershipRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	// 只有当前所有者可以转让
	if enterprise.CreatorID != userID {
		return models.NewJsonResponse(403, "只有企业创建者可以转让所有权", nil), nil
	}

	// 支持通过用户名或用户ID查找新所有者
	newOwnerID := req.NewOwnerID
	if newOwnerID == "" && req.NewOwnerName != "" {
		// 通过成员列表查找用户名对应的用户ID
		members, err := s.factory.EnterpriseMember().ListByEnterpriseID(ctx, req.EnterpriseID, 0, 0)
		if err != nil {
			return models.NewJsonResponse(500, "查询成员失败", nil), err
		}
		for _, m := range members {
			user, _ := s.factory.User().GetByID(ctx, m.UserID)
			if user != nil && (user.Name == req.NewOwnerName || user.UserName == req.NewOwnerName) {
				newOwnerID = m.UserID
				break
			}
		}
		if newOwnerID == "" {
			return models.NewJsonResponse(400, "未找到该用户名对应的成员", nil), nil
		}
	}
	if newOwnerID == "" {
		return models.NewJsonResponse(400, "请指定新所有者", nil), nil
	}

	// 验证新所有者是企业成员
	newOwnerMember, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, newOwnerID)
	if err != nil || newOwnerMember == nil || newOwnerMember.Status != 0 {
		return models.NewJsonResponse(400, "新所有者不是该企业的活跃成员", nil), nil
	}

	// 将新所有者设为管理员角色
	adminRole, err := s.factory.EnterpriseRole().GetAdminByEnterpriseID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "获取管理员角色失败", nil), err
	}
	newOwnerMember.RoleID = adminRole.ID
	if err := s.factory.EnterpriseMember().Update(ctx, newOwnerMember); err != nil {
		return models.NewJsonResponse(500, "更新成员角色失败", nil), err
	}

	// 转让所有权
	enterprise.CreatorID = newOwnerID
	if err := s.factory.Enterprise().Update(ctx, enterprise); err != nil {
		return models.NewJsonResponse(500, "转让失败", nil), err
	}

	// 旧创建者降级为普通成员
	defaultRole, _ := s.factory.EnterpriseRole().GetDefaultByEnterpriseID(ctx, req.EnterpriseID)
	if defaultRole != nil {
		oldOwnerMember, _ := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
		if oldOwnerMember != nil {
			oldOwnerMember.RoleID = defaultRole.ID
			s.factory.EnterpriseMember().Update(ctx, oldOwnerMember)
		}
	}

	return models.NewJsonResponse(200, "转让成功", nil), nil
}

// DissolveEnterprise 解散企业
func (s *EnterpriseService) DissolveEnterprise(req *request.DissolveEnterpriseRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	if enterprise.CreatorID != userID {
		return models.NewJsonResponse(403, "只有企业创建者可以解散企业", nil), nil
	}

	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)

		if err := txFactory.EnterpriseSharedFile().DeleteByEnterpriseID(ctx, req.EnterpriseID); err != nil {
			return err
		}
		if err := txFactory.EnterpriseSharedPath().DeleteByEnterpriseID(ctx, req.EnterpriseID); err != nil {
			return err
		}

		roles, _ := txFactory.EnterpriseRole().ListByEnterpriseID(ctx, req.EnterpriseID)
		for _, role := range roles {
			if err := txFactory.EnterpriseRolePower().DeleteByRoleID(ctx, role.ID); err != nil {
				return err
			}
			if err := txFactory.EnterpriseRole().Delete(ctx, role.ID); err != nil {
				return err
			}
		}

		members, _ := txFactory.EnterpriseMember().ListByEnterpriseID(ctx, req.EnterpriseID, 0, 0)
		for _, m := range members {
			if err := txFactory.EnterpriseMember().Delete(ctx, m.ID); err != nil {
				return err
			}
		}

		invites, _ := txFactory.EnterpriseInvite().ListByEnterpriseID(ctx, req.EnterpriseID, 0, 0)
		for _, inv := range invites {
			inv.Status = 3
			if err := txFactory.EnterpriseInvite().Update(ctx, inv); err != nil {
				return err
			}
		}

		return txFactory.Enterprise().Delete(ctx, req.EnterpriseID)
	})

	if err != nil {
		return models.NewJsonResponse(500, "解散企业失败", nil), err
	}

	return models.NewJsonResponse(200, "企业已解散", nil), nil
}

// ToggleEnterpriseState 启用/禁用企业
func (s *EnterpriseService) ToggleEnterpriseState(req *request.ToggleEnterpriseStateRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	enterprise.State = req.State
	if err := s.factory.Enterprise().Update(ctx, enterprise); err != nil {
		return models.NewJsonResponse(500, "操作失败", nil), err
	}

	stateText := "启用"
	if req.State == 1 {
		stateText = "禁用"
	}

	return models.NewJsonResponse(200, stateText+"成功", nil), nil
}

// SetEnterpriseQuota 设置企业存储配额
func (s *EnterpriseService) SetEnterpriseQuota(req *request.SetEnterpriseQuotaRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := s.checkAdmin(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "需要管理员权限", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	if req.SpaceUnlimited {
		// 设置为无限空间
		enterprise.SpaceUnlimited = true
		enterprise.Space = 0
		enterprise.FreeSpace = 0
	} else {
		// 计算已用空间
		usedSpace, _ := s.factory.EnterpriseSharedFile().SumSizeByEnterpriseID(ctx, req.EnterpriseID)

		enterprise.SpaceUnlimited = false
		enterprise.Space = req.Space
		if req.Space > 0 {
			enterprise.FreeSpace = req.Space - usedSpace
			if enterprise.FreeSpace < 0 {
				enterprise.FreeSpace = 0
			}
		} else {
			enterprise.FreeSpace = 0
		}
	}

	if err := s.factory.Enterprise().Update(ctx, enterprise); err != nil {
		return models.NewJsonResponse(500, "设置失败", nil), err
	}

	return models.NewJsonResponse(200, "设置成功", map[string]interface{}{
		"space":           enterprise.Space,
		"free_space":      enterprise.FreeSpace,
		"space_unlimited": enterprise.SpaceUnlimited,
	}), nil
}

// GetEnterpriseAuditLogs 查询企业审计日志
func (s *EnterpriseService) GetEnterpriseAuditLogs(req *request.EnterpriseAuditListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证用户是企业成员
	if err := s.checkMember(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, "无权访问", nil), err
	}

	query := &repository.AuditLogQuery{
		EnterpriseID: req.EnterpriseID,
		UserID:       "",
		Action:       req.Action,
		Keyword:      req.Keyword,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}

	logs, total, err := s.factory.AuditLog().ListByCondition(ctx, query)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"list":     logs,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}), nil
}

// ExportEnterpriseAuditLogs 导出企业审计日志为CSV
func (s *EnterpriseService) ExportEnterpriseAuditLogs(enterpriseID, action, keyword, startTime, endTime, userID string) ([]byte, error) {
	ctx := context.Background()

	if err := s.checkMember(ctx, enterpriseID, userID); err != nil {
		return nil, fmt.Errorf("无权访问")
	}

	query := &repository.AuditLogQuery{
		EnterpriseID: enterpriseID,
		Action:       action,
		Keyword:      keyword,
		StartTime:    startTime,
		EndTime:      endTime,
		Page:         1,
		PageSize:     10000,
	}

	logs, _, err := s.factory.AuditLog().ListByCondition(ctx, query)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"ID", "用户ID", "用户名", "操作", "目标类型", "目标名称", "详情", "IP", "时间"})
	for _, l := range logs {
		writer.Write([]string{l.ID, l.UserID, l.UserName, l.Action, l.TargetType, l.TargetName, l.Detail, l.IP, l.CreatedAt.Format("2006-01-02 15:04:05")})
	}
	writer.Flush()

	return buf.Bytes(), nil
}

// checkAdmin 检查用户是否为企业管理员
func (s *EnterpriseService) checkAdmin(ctx context.Context, enterpriseID, userID string) error {
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	if err != nil || member == nil {
		return fmt.Errorf("not a member")
	}
	if member.Status != 0 {
		return fmt.Errorf("member inactive")
	}
	role, err := s.factory.EnterpriseRole().GetByID(ctx, member.RoleID)
	if err != nil || role == nil || role.IsAdmin != 1 {
		return fmt.Errorf("not admin")
	}
	return nil
}

// checkMember 检查用户是否为企业成员
func (s *EnterpriseService) checkMember(ctx context.Context, enterpriseID, userID string) error {
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	if err != nil || member == nil {
		return fmt.Errorf("not a member")
	}
	if member.Status != 0 {
		return fmt.Errorf("member inactive")
	}
	return nil
}

// GetEnterpriseMember 获取企业成员信息（通过 member ID）
func (s *EnterpriseService) GetEnterpriseMember(ctx context.Context, memberID string) (*models.EnterpriseMember, error) {
	return s.factory.EnterpriseMember().GetByID(ctx, memberID)
}

// GetAllPowers 获取企业可用权限列表（仅返回 enterprise: 前缀的权限）
func (s *EnterpriseService) GetAllPowers() (*models.JsonResponse, error) {
	ctx := context.Background()
	allPowers, err := s.factory.Power().List(ctx, 0, 1000)
	if err != nil {
		return models.NewJsonResponse(500, "查询失败", nil), err
	}
	var enterprisePowers []*models.Power
	for _, p := range allPowers {
		if len(p.Characteristic) >= 11 && p.Characteristic[:11] == "enterprise:" {
			enterprisePowers = append(enterprisePowers, p)
		}
	}
	return models.NewJsonResponse(200, "ok", enterprisePowers), nil
}

// expireTime 生成过期时间
func expireTime(days int) custom_type.JsonTime {
	now := custom_type.Now()
	return now.Add(time.Duration(days) * 24 * time.Hour)
}

// generateInviteCode 生成6位随机邀请码
func generateInviteCode() (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}
	return string(code), nil
}
