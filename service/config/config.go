package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// configDef 配置定义（用于注册所有已知的系统配置项）
type configDef struct {
	Key          string
	DefaultValue string
}

// knownConfigs 已注册的系统配置项
var knownConfigs = []configDef{
	{Key: util.ServerPortKey, DefaultValue: util.DefaultServerPort},
	{Key: util.ConfigKeyAuditRetentionDays, DefaultValue: util.DefaultAuditRetentionDays},
	{Key: util.ConfigKeyAlertRetentionDays, DefaultValue: util.DefaultAlertRetentionDays},
	{Key: util.ConfigKeyJwtSecret, DefaultValue: ""},
}

// defaultMap 快速查找默认值
func defaultMap() map[string]string {
	m := make(map[string]string, len(knownConfigs))
	for _, d := range knownConfigs {
		m[d.Key] = d.DefaultValue
	}
	return m
}

// Get 获取所有系统设置
func Get(c *gin.Context) {
	defaults := defaultMap()
	items := make([]Item, 0, len(knownConfigs))

	for _, def := range knownConfigs {
		value := models.GetConfig(def.Key, def.DefaultValue)
		items = append(items, Item{
			Key:          def.Key,
			Value:        value,
			DefaultValue: defaults[def.Key],
		})
	}

	util.ResponseOkWithData(c, &GetResponse{Items: items})
}

// Update 更新系统设置
func Update(c *gin.Context, in *UpdateRequest) {
	defaults := defaultMap()
	for _, item := range in.Items {
		// 只允许更新已注册的配置项
		if _, ok := defaults[item.Key]; !ok {
			util.ResponseError(c, fmt.Errorf("unknown config key: %s", item.Key))
			return
		}
		if err := models.SetConfig(item.Key, item.Value); err != nil {
			logger.Error("[config] update config error.", zap.String("key", item.Key), zap.Error(err))
			util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
			return
		}
		// JWT 密钥变更时同步更新内存中的密钥（重启后也会从 DB 重新加载）
		if item.Key == util.ConfigKeyJwtSecret && item.Value != "" {
			util.JwtKey = item.Value
			logger.Info("[config] JWT secret key updated")
		}
	}
	util.ResponseOk(c)
}
