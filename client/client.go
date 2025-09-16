package client

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"time"
)

// Post 带鉴权的 post 请求
//
//	path: 请求路径
//	data: 请求数据
func Post(path string, data any) ([]byte, error) {
	// 鉴权
	ub := &models.UserBasic{Level: models.UserLevelRoot}
	if err := ub.First(); err != nil {
		logger.Error("[sys] get user basic by level error.", zap.Error(err))
		return nil, err
	}
	token, err := (&util.UserClaim{
		Username: ub.Username,
		Level:    ub.Level,
	}).GenerateToken(time.Now().Add(time.Hour * 24).Unix())
	if err != nil {
		logger.Error("[sys] generate token error.", zap.Error(err))
		return nil, err
	}
	// header
	headers := map[string]string{
		"Authorization": token,
	}
	headerBytes, err := jsoniter.Marshal(headers)
	if err != nil {
		logger.Error("[sys] marshal header error.", zap.Error(err))
		return nil, err
	}

	// 获取服务端口
	serverPort, err := (&models.ConfigBasic{}).GetServerPort()
	if err != nil {
		logger.Error("[db] get server port error.", zap.Error(err))
		return nil, err
	}

	// 发送请求
	return netutil.HttpPost(fmt.Sprintf("http://127.0.0.1%s%s", serverPort, path), data, headerBytes...)
}
