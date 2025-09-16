package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/util"
)

// LoginAuthCheck 登录信息认证
func LoginAuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		userClaim, err := util.AnalyzeToken(token)
		if err != nil {
			c.Abort()
			util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
		} else {
			if userClaim.Username == "" {
				c.Abort()
				util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
			}
			c.Set("UserClaim", userClaim)
			c.Next()
		}
	}
}
