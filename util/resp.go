package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	CodeOk              = 200   // 正常
	CodeErr             = -1    // 通用异常
	CodeErrForceModify  = 60201 // 强制修改时使用
	CodeErrPlatform     = 60202 // 第三方调用异常
	CodeErrAuth         = 60400 // 登录过期，请重新登录
	CodeErrParam        = 60403 // 参数异常
	CodeErrDataNotExist = 60404 // 数据不存在
	CodeErrDB           = 60500 // 数据库异常
)

const (
	MsgOk                    = "Success"
	MsgErr                   = "系统异常"
	MsgErrAuth               = "登录过期，请重新登录"
	MsgErrParam              = "参数异常"
	MsgErrDB                 = "数据库异常"
	MsgErrNet                = "网络异常"
	MsgErrUsernameOrPassword = "用户名或密码错误"
	MsgErrOldPasswordWrong   = "旧密码错误"
	MsgErrNameExist          = "名称已存在"
	MsgErrDataNotExist       = "暂无数据"
)

type BaseResponse struct {
	Code int    `json:"code"` // 错误码
	Msg  string `json:"msg"`  // 提示文本
	Data any    `json:"data"` // 数据
}

// ResponseWithData 发送携带数据的响应
func ResponseWithData(c *gin.Context, code int, msg string, data any) {
	c.JSON(http.StatusOK, &BaseResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// ResponseMsg 未携带数据的响应
func ResponseMsg(c *gin.Context, code int, msg string) {
	ResponseWithData(c, code, msg, nil)
}

// ResponseOkWithData 携带数据的成功响应
func ResponseOkWithData(c *gin.Context, data any) {
	ResponseWithData(c, CodeOk, MsgOk, data)
}

// ResponseOkWithList 数据列表的成功响应
//
// list 数据列表
// count 数据总数
func ResponseOkWithList(c *gin.Context, list any, count ...int64) {
	if len(count) > 0 {
		ResponseWithData(c, CodeOk, MsgOk, gin.H{
			"list":  list,
			"count": count[0],
		})
		return
	}
	ResponseWithData(c, CodeOk, MsgOk, list)
}

// ResponseOk 成功的响应
func ResponseOk(c *gin.Context) {
	ResponseWithData(c, CodeOk, MsgOk, nil)
}

// ResponseError 错误响应
func ResponseError(c *gin.Context, err error) {
	ResponseWithData(c, CodeErr, err.Error(), nil)
}

// ResponseFile 简单文件返回
//
// data 文件数据
// fileName 文件名，例如：xxx.json
func ResponseFile(c *gin.Context, data []byte, fileName string) {
	// 设置 HTTP 头部
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	// 发送文件数据
	c.Data(http.StatusOK, "application/octet-stream", data)
}
