package dashboard

import (
	"errors"

	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/client"
	serviceDashboard "github.com/up-zero/my-proxy/service/dashboard"
	"github.com/up-zero/my-proxy/util"
)

// Overview 获取仪表盘总览数据
func Overview() (*serviceDashboard.OverviewResponse, error) {
	res, err := netutil.ParseResponse[OverviewResponse](client.Get("/api/v1/dashboard/overview"))
	if err != nil {
		return nil, err
	}
	if res.Code != util.CodeOk {
		return nil, errors.New(res.Msg)
	}
	return res.Data, nil
}
