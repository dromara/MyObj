package middleware

import (
	"myobj/src/config"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// AdminVerify 管理员权限验证中间件
// 检查用户是否为管理员（通过配置中的 admin_group_id）
func AdminVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, exist := c.Get("userLogin")
		if !exist {
			c.JSON(401, models.NewJsonResponse(401, "用户未登录", nil))
			c.Abort()
			return
		}
		loginInfo, ok := userLogin.(response.UserLoginResponse)
		if !ok {
			c.JSON(500, models.NewJsonResponse(500, "内部错误", nil))
			c.Abort()
			return
		}

		// 检查用户组ID是否为管理员组
		// AdminGroupID 为 0 表示未配置管理员组，此时拒绝所有管理员请求
		if loginInfo.User == nil || config.CONFIG.Auth.AdminGroupID == 0 || loginInfo.User.GroupID != config.CONFIG.Auth.AdminGroupID {
			c.JSON(403, models.NewJsonResponse(403, "需要管理员权限", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

