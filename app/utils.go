package app

import (
	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/structutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// BindH 绑定请求参数
func BindH[T any](fn func(*gin.Context, *T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		in, err := structutil.NewWithDefaults[T]()
		if err != nil {
			logger.Error("[sys] request param bind error.",
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
				zap.Error(err))
			util.ResponseMsg(c, util.CodeErrParam, util.MsgErrParam)
			return
		}

		switch c.Request.Method {
		case "GET", "DELETE":
			err = c.ShouldBindQuery(in)
		case "POST", "PUT":
			err = c.ShouldBindJSON(in)
		default:
			err = c.ShouldBind(in)
		}
		if err != nil {
			logger.Error("[sys] request param bind error.",
				zap.String("path", c.FullPath()),
				zap.String("method", c.Request.Method),
				zap.Error(err))
			util.ResponseMsg(c, util.CodeErrParam, util.MsgErrParam)
			return
		}
		fn(c, in)
	}
}

// BindSliceH 绑定请求参数（切片）
func BindSliceH[T any](fn func(*gin.Context, []*T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in = make([]*T, 0)
		if err := c.ShouldBindJSON(&in); err != nil {
			logger.Error("[sys] request param bind error",
				zap.String("path", c.FullPath()),
				zap.Error(err))
			util.ResponseMsg(c, util.CodeErrParam, util.MsgErrParam)
			return
		}
		fn(c, in)
	}
}
