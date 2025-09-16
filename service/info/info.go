package info

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

func Info(c *gin.Context) {
	reply := &Reply{
		Version:   fmt.Sprintf("%s %s", util.AppName, util.AppVersion),
		Addresses: nil,
		Username:  "",
		Password:  "",
	}
	// 地址
	ips, err := netutil.Ipv4sLocal()
	if err != nil {
		logger.Error("[gotool] get ipv4 error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	serverPort, err := (&models.ConfigBasic{}).GetServerPort()
	if err != nil {
		logger.Error("[db] get server port error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	for _, ip := range ips {
		// http://127.0.0.1:12321
		reply.Addresses = append(reply.Addresses, fmt.Sprintf("http://%s%s", ip, serverPort))
	}

	// 用户信息
	ub := &models.UserBasic{Level: models.UserLevelRoot}
	if err := ub.First(); err != nil {
		logger.Error("[db] get user basic error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	reply.Username = ub.Username
	reply.Password = ub.Password

	util.ResponseOkWithData(c, reply)
}
