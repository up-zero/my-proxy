package models

import (
	"github.com/up-zero/my-proxy/util"
)

type ConfigBasic struct {
	Key       string `json:"key"`                                                       // 键
	Value     string `json:"value"`                                                     // 值
	CreatedAt int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

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

// GetServerPort 获取服务端口
func (table *ConfigBasic) GetServerPort() (string, error) {
	if err := DB.Model(table).First(table, &ConfigBasic{Key: util.ServerPortKey}).Error; err != nil {
		return "", err
	}
	return table.Value, nil
}

// SaveServerPort 保存服务端口
func (table *ConfigBasic) SaveServerPort(port string) error {
	table.Key = util.ServerPortKey
	table.Value = port
	return DB.FirstOrCreate(table, &ConfigBasic{Key: util.ServerPortKey}).Error
}
