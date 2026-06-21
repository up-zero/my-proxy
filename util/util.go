package util

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	AppName                 = "my-proxy"
	AppVersion              = "1.1.0"
	DateTimeWithMilliLayout = "2006-01-02 15:04:05.000"
	JwtKey                  = "LnGvpbI59mPaSxA3"

	// ---- 系统设置 KEY ----

	// ServerPortKey 服务端口 KEY
	ServerPortKey = "SERVER_PORT_KEY"
	// ConfigKeyAuditRetentionDays 日志审计存储时长（天）
	ConfigKeyAuditRetentionDays = "AUDIT_RETENTION_DAYS"
	// ConfigKeyAlertRetentionDays 告警通知存储时长（天）
	ConfigKeyAlertRetentionDays = "ALERT_RETENTION_DAYS"

	// ---- 系统设置默认值 ----

	DefaultAuditRetentionDays = "90"    // 日志审计默认 90 天
	DefaultAlertRetentionDays = "90"    // 告警通知默认 90 天
	DefaultServerPort         = "12312" // 服务端口默认值
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
