package middleware

import (
	"myobj/src/core/domain/response"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// AdminVerify 管理员权限验证中间件
// 检查用户是否为管理员（group_id = 1）
func AdminVerify() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, exist := c.Get("userLogin")
		if !exist {
			c.JSON(401, models.NewJsonResponse(401, "用户未登录", nil))
			c.Abort()
			return
		}
		loginInfo := userLogin.(response.UserLoginResponse)
		
		// 检查用户组ID是否为1（管理员组）
		if loginInfo.User == nil || loginInfo.User.GroupID != 1 {
			c.JSON(403, models.NewJsonResponse(403, "需要管理员权限", nil))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

