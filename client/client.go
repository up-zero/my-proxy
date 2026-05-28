package client

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/up-zero/gotool/netutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func authHeaders() (map[string]string, error) {
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
	return map[string]string{
		"Authorization": token,
	}, nil
}

func serverURL(path string) (string, error) {
	// 获取服务端口
	serverPort, err := (&models.ConfigBasic{}).GetServerPort()
	if err != nil {
		logger.Error("[db] get server port error.", zap.Error(err))
		return "", err
	}
	return fmt.Sprintf("http://127.0.0.1%s%s", serverPort, path), nil
}

// Post 带鉴权的 post 请求
//
//	path: 请求路径
//	data: 请求数据
func Post(path string, data any) ([]byte, error) {
	headers, err := authHeaders()
	if err != nil {
		return nil, err
	}
	headerBytes, err := jsoniter.Marshal(headers)
	if err != nil {
		logger.Error("[sys] marshal header error.", zap.Error(err))
		return nil, err
	}
	url, err := serverURL(path)
	if err != nil {
		return nil, err
	}

	// 发送请求
	return netutil.HttpPost(url, data, headerBytes...)
}

// Get 带鉴权的 get 请求
func Get(path string) ([]byte, error) {
	headers, err := authHeaders()
	if err != nil {
		return nil, err
	}
	url, err := serverURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest && len(body) == 0 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	return body, nil
}
