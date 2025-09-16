package proxy

import (
	"github.com/up-zero/my-proxy/service/info"
	"github.com/up-zero/my-proxy/service/serve"
)

type StatusResponse struct {
	Code int                `json:"code"` // code
	Msg  string             `json:"msg"`  // 消息
	Data []*serve.ProxyTask `json:"data"` // 列表数据
}

type Response struct {
	Code int         `json:"code"` // code
	Msg  string      `json:"msg"`  // 消息
	Data *info.Reply `json:"data"` // 数据
}

type EmptyResponse struct {
	Code int    `json:"code"` // code
	Msg  string `json:"msg"`  // 消息
}
