package models

import (
	"strconv"
	"strings"
	"sync"

	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

type ConfigBasic struct {
	Key       string `json:"key"`                                                       // 键
	Value     string `json:"value"`                                                     // 值
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

// configCache 内存缓存，减少磁盘 IO
// key: 配置 key，value: 配置 value（string）
var configCache sync.Map

func (table *ConfigBasic) TableName() string {
	return "config_basic"
}

func (table *ConfigBasic) First() error {
	tx := DB.Model(table)
	if table.Key != "" {
		tx = tx.Where("key = ?", table.Key)
	}
	err := tx.First(table).Error
	return err
}

// InitConfigCache 启动时调用：将 config_basic 全表加载进内存缓存
func InitConfigCache() {
	var rows []ConfigBasic
	if err := DB.Model(new(ConfigBasic)).Find(&rows).Error; err != nil {
		logger.Error("[db] init config cache error.", zap.Error(err))
		return
	}
	for _, row := range rows {
		// 清理服务端口配置中的前导冒号
		if row.Key == util.ServerPortKey {
			cleanedPort := row.Value
			for strings.HasPrefix(cleanedPort, ":") {
				cleanedPort = strings.TrimPrefix(cleanedPort, ":")
			}
			if cleanedPort != row.Value {
				logger.Info("[config] auto-fixing port config", zap.String("old", row.Value), zap.String("new", cleanedPort))
				// 更新数据库和缓存
				if err := SetConfig(util.ServerPortKey, cleanedPort); err != nil {
					logger.Error("[config] fix port config error", zap.Error(err))
				}
				configCache.Store(row.Key, cleanedPort)
				continue
			}
		}
		configCache.Store(row.Key, row.Value)
	}
	logger.Info("[config] config cache initialized", zap.Int("count", len(rows)))
}

// GetConfig 获取配置值（优先读缓存，未命中则查库并回填缓存）
func GetConfig(key, defaultValue string) string {
	// 先查缓存
	if v, ok := configCache.Load(key); ok {
		return v.(string)
	}
	// 查库
	row := &ConfigBasic{Key: key}
	if err := DB.Model(new(ConfigBasic)).Where("key = ?", key).First(row).Error; err != nil {
		configCache.Store(key, defaultValue)
		return defaultValue
	}
	// 回填缓存
	configCache.Store(key, row.Value)
	return row.Value
}

// GetConfigInt 获取 int 类型配置
func GetConfigInt(key string, defaultValue int) int {
	raw := GetConfig(key, strconv.Itoa(defaultValue))
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

// SetConfig 设置配置（先落库，再更新缓存）
func SetConfig(key, value string) error {
	row := &ConfigBasic{Key: key}
	result := DB.Model(new(ConfigBasic)).Where("key = ?", key).First(row)
	if result.Error != nil {
		// 记录不存在，新增
		row.Value = value
		if err := DB.Create(row).Error; err != nil {
			return err
		}
	} else {
		// 记录已存在，更新
		if err := DB.Model(new(ConfigBasic)).Where("key = ?", key).Update("value", value).Error; err != nil {
			return err
		}
	}
	// 更新缓存
	configCache.Store(key, value)
	return nil
}

// GetServerPort 获取服务端口
func (table *ConfigBasic) GetServerPort() string {
	if err := DB.Model(table).First(table, &ConfigBasic{Key: util.ServerPortKey}).Error; err != nil {
		logger.Warn("[config] get server port error", zap.Error(err))
		return util.DefaultServerPort
	}
	port := table.Value
	for strings.HasPrefix(port, ":") {
		port = strings.TrimPrefix(port, ":")
	}
	return port
}

// SaveServerPort 保存服务端口
func (table *ConfigBasic) SaveServerPort(port string) error {
	return SetConfig(util.ServerPortKey, port)
}
