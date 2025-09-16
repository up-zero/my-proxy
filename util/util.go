package util

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	AppName                 = "my-proxy"
	AppVersion              = "1.0.0"
	DateTimeWithMilliLayout = "2006-01-02 15:04:05.000"
	JwtKey                  = "LnGvpbI59mPaSxA3"

	// ServerPortKey 服务端口 KEY
	ServerPortKey = "SERVER_PORT_KEY"
)

// GetDbPath 获取数据库路径
func GetDbPath() (string, error) {
	dbDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDBPath := filepath.Join(dbDir, AppName, "my-proxy.db")
	if err := os.MkdirAll(filepath.Dir(appDBPath), os.ModePerm); err != nil {
		return "", err
	}
	return appDBPath, nil
}

// GetLogPath 获取日志路径
func GetLogPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("[sys] get user config dir error: %v", err)
		return ""
	}
	appLogPath := filepath.Join(configDir, AppName, "log", "my-proxy.log")
	if err := os.MkdirAll(filepath.Dir(appLogPath), os.ModePerm); err != nil {
		return ""
	}
	return appLogPath
}
