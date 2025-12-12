package middleware

import (
	"myobj/src/core/domain/response"
	"myobj/src/pkg/models"

	"github.com/gin-gonic/gin"
)

// PowerVerify 权限验证中间件
func PowerVerify(power string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, exist := c.Get("userLogin")
		if !exist {
			c.JSON(401, models.NewJsonResponse(401, "用户未登录", nil))
			c.Abort()
			return
		}
		loginInfo := userLogin.(response.UserLoginResponse)
		for _, p := range loginInfo.Power {
			if p.Characteristic == power {
				c.Next()
				return
			}
		}
		c.JSON(401, models.NewJsonResponse(401, "用户无权限", nil))
		c.Abort()
		return
	}
}
