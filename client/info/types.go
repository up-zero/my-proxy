package info

import "github.com/up-zero/my-proxy/service/info"

type Response struct {
	Code int         `json:"code"` // code
	Msg  string      `json:"msg"`  // 消息
	Data *info.Reply `json:"data"` // 数据
}
