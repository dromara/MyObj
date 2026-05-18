package middleware

import (
	"context"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/models"
	"myobj/src/pkg/repository"

	"github.com/gin-gonic/gin"
)

// EnterpriseMiddleware 企业上下文中间件
type EnterpriseMiddleware struct {
	enterpriseMemberRepo    repository.EnterpriseMemberRepository
	enterpriseRoleRepo      repository.EnterpriseRoleRepository
	enterpriseRolePowerRepo repository.EnterpriseRolePowerRepository
	powerRepo               repository.PowerRepository
}

// NewEnterpriseMiddleware 创建企业中间件
func NewEnterpriseMiddleware(
	enterpriseMemberRepo repository.EnterpriseMemberRepository,
	enterpriseRoleRepo repository.EnterpriseRoleRepository,
	enterpriseRolePowerRepo repository.EnterpriseRolePowerRepository,
	powerRepo repository.PowerRepository,
) *EnterpriseMiddleware {
	return &EnterpriseMiddleware{
		enterpriseMemberRepo:    enterpriseMemberRepo,
		enterpriseRoleRepo:      enterpriseRoleRepo,
		enterpriseRolePowerRepo: enterpriseRolePowerRepo,
		powerRepo:               powerRepo,
	}
}

// Verify 企业上下文加载中间件
// 如果用户已切换到企业（CurrentEnterpriseID非空），加载企业成员、角色、权限到gin.Context
// 如果未切换到企业，跳过（不中断请求）
func (m *EnterpriseMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLoginVal, exists := c.Get("userLogin")
		if !exists {
			c.Next()
			return
		}
		userLogin := userLoginVal.(response.UserLoginResponse)
		if userLogin.User == nil || userLogin.User.CurrentEnterpriseID == "" {
			c.Next()
			return
		}

		ctx := context.Background()
		enterpriseID := userLogin.User.CurrentEnterpriseID

		// 加载企业成员记录，验证 status=0（活跃）
		member, err := m.enterpriseMemberRepo.GetByEnterpriseAndUser(ctx, enterpriseID, userLogin.User.ID)
		if err != nil || member == nil || member.Status != 0 {
			c.JSON(403, models.NewJsonResponse(403, "企业成员验证失败", nil))
			c.Abort()
			return
		}

		// 加载企业角色
		role, err := m.enterpriseRoleRepo.GetByID(ctx, member.RoleID)
		if err != nil || role == nil {
			c.JSON(403, models.NewJsonResponse(403, "企业角色验证失败", nil))
			c.Abort()
			return
		}

		// 加载企业角色权限
		rolePowers, err := m.enterpriseRolePowerRepo.GetByRoleID(ctx, role.ID)
		if err != nil {
			c.JSON(403, models.NewJsonResponse(403, "企业权限加载失败", nil))
			c.Abort()
			return
		}

		var enterprisePowers []*models.Power
		for _, rp := range rolePowers {
			power, err := m.powerRepo.GetByID(ctx, rp.PowerID)
			if err == nil && power != nil {
				enterprisePowers = append(enterprisePowers, power)
			}
		}

		// 设置企业上下文
		c.Set("enterpriseID", enterpriseID)
		c.Set("enterpriseMember", member)
		c.Set("enterpriseRole", role)
		c.Set("enterprisePowers", enterprisePowers)

		c.Next()
	}
}

// PowerVerify 企业权限验证中间件
// 检查用户的企业角色是否拥有指定权限
// 必须在 EnterpriseVerify 之后使用
func (m *EnterpriseMiddleware) PowerVerify(power string) gin.HandlerFunc {
	return func(c *gin.Context) {
		enterpriseID, exists := c.Get("enterpriseID")
		if !exists || enterpriseID == "" {
			c.JSON(403, models.NewJsonResponse(403, "请先切换到企业空间", nil))
			c.Abort()
			return
		}

		powersVal, exists := c.Get("enterprisePowers")
		if !exists {
			c.JSON(403, models.NewJsonResponse(403, "无企业权限", nil))
			c.Abort()
			return
		}
		powers := powersVal.([]*models.Power)

		for _, p := range powers {
			if p.Characteristic == power {
				c.Next()
				return
			}
		}

		c.JSON(403, models.NewJsonResponse(403, "无此企业操作权限", nil))
		c.Abort()
	}
}

// AdminVerify 企业管理员验证中间件
// 检查用户在当前企业中是否为管理员
// 必须在 EnterpriseVerify 之后使用
func (m *EnterpriseMiddleware) AdminVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("enterpriseRole")
		if !exists {
			c.JSON(403, models.NewJsonResponse(403, "请先切换到企业空间", nil))
			c.Abort()
			return
		}
		role := roleVal.(*models.EnterpriseRole)
		if role.IsAdmin != 1 {
			c.JSON(403, models.NewJsonResponse(403, "需要企业管理员权限", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
