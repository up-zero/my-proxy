package proxy

import (
	"errors"
	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/client"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/proxy"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
)

// Status 获取代理状态
func Status(name string) ([]*serve.ProxyTask, error) {
	res, err := netutil.ParseResponse[StatusResponse](client.Post("/api/v1/proxy/status", map[string]any{
		"name": name,
	}))
	if err != nil {
		return nil, err
	}
	if res.Code != util.CodeOk {
		return nil, errors.New(res.Msg)
	}
	return res.Data, nil
}

// Create 创建代理
func Create(in *proxy.CreateRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/create", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}

// GetDetailByName 通过名称获取代理详情
func GetDetailByName(name string) (*models.ProxyBasic, error) {
	pb := new(models.ProxyBasic)
	err := models.DB.Model(pb).Where("name = ?", name).First(pb).Error
	return pb, err
}

// Edit 编辑代理
func Edit(in *proxy.EditRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/edit", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}

// Delete 删除代理
func Delete(in *proxy.DeleteRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/delete", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}

// Start 启动代理
func Start(in *proxy.StartRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/start", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}

// Stop 停止代理
func Stop(in *proxy.StopRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/stop", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}

// Restart 重启代理
func Restart(in *proxy.RestartRequest) error {
	res, err := netutil.ParseResponse[EmptyResponse](client.Post("/api/v1/proxy/restart", in))
	if err != nil {
		return err
	}
	if res.Code != util.CodeOk {
		return errors.New(res.Msg)
	}
	return nil
}
