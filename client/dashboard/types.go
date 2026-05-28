package dashboard

import serviceDashboard "github.com/up-zero/my-proxy/service/dashboard"

type OverviewResponse struct {
	Code int                                `json:"code"`
	Msg  string                             `json:"msg"`
	Data *serviceDashboard.OverviewResponse `json:"data"`
}
