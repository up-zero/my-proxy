package serve

import (
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"go.uber.org/zap"
)

// NewProxy 初始化代理
func NewProxy() {
	proxies, err := (&models.ProxyBasic{}).All()
	if err != nil {
		logger.Error("[sys] get proxy list error", zap.Error(err))
		return
	}

	for _, proxy := range proxies {
		task := &ProxyTask{
			ProxyBasic: *proxy,
		}
		if err := task.Start(); err != nil {
			logger.Error("[sys] proxy task start error", zap.Error(err))
		}
	}
}
