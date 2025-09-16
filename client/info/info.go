package info

import (
	"errors"
	"github.com/up-zero/gotool"
	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/client"
	"github.com/up-zero/my-proxy/service/info"
	"github.com/up-zero/my-proxy/util"
)

// Info 获取服务信息
func Info() (*info.Reply, error) {
	res, err := netutil.ParseResponse[Response](client.Post("/api/v1/info", gotool.EmptyMapBytes))
	if err != nil {
		return nil, err
	}
	if res.Code != util.CodeOk {
		return nil, errors.New(res.Msg)
	}
	return res.Data, nil
}
