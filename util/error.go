package util

import "errors"

var (
	// ErrNameExists 名称已存在
	ErrNameExists = errors.New("name already exists")
	// ErrListenPortExists 监听端口已存在
	ErrListenPortExists = errors.New("listen port already exists")
	// ErrProxyNotExist 代理不存在
	ErrProxyNotExist = errors.New("proxy not exists")
)
