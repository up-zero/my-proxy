package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"strings"
)

func getTokenFromRequest(c *gin.Context) string {
	tokens := []string{
		c.GetHeader("Authorization"),
		c.Query("token"),
		c.Query("access_token"),
		c.Query("authorization"),
	}

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			return token
		}
	}

	return ""
}

// LoginAuthCheck 登录信息认证
func LoginAuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromRequest(c)
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

// AdminAuthCheck 超管信息认证
func AdminAuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromRequest(c)
		userClaim, err := util.AnalyzeToken(token)
		if err != nil {
			c.Abort()
			util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
		} else {
			if userClaim.Level != models.UserLevelRoot {
				c.Abort()
				util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
			}
			c.Set("UserClaim", userClaim)
			c.Next()
		}
	}
}
