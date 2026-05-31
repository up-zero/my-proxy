package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/util"
)

func Language() gin.HandlerFunc {
	return func(c *gin.Context) {
		util.SetRequestLanguage(c, util.DetectRequestLanguage(c))
		c.Next()
	}
}
