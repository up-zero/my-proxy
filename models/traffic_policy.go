package models

const (
	TrafficPolicyStatusEnabled  = "ENABLED"
	TrafficPolicyStatusDisabled = "DISABLED"
)

const (
	OverLimitActionSlowdown = "SLOWDOWN"
	OverLimitActionAlert    = "ALERT"
)

type TrafficPolicy struct {
	Uuid                string   `json:"uuid"`
	Name                string   `json:"name"`
	ScopeType           string   `json:"scope_type"`          // 作用范围类型：ALL / TAG / PROXY
	ScopeValue          string   `json:"scope_value"`         // 作用范围值：标签 uuid / 代理 uuid
	ScopeName           string   `gorm:"-" json:"scope_name"` // 作用范围名称
	OutboundLimit       string   `json:"outbound_limit"`      // 出站上限，例如：50MB/s
	MaxConnections      string   `json:"max_connections"`     // 并发连接，例如：300
	PeriodQuota         string   `json:"period_quota"`        // 周期配额，例如：800GB/月
	QuotaUsed           string   `json:"quota_used"`          // 配额已用额度，例如：536GB
	OverLimitAction     string   `json:"over_limit_action"`   // 超额动作，逗号分隔：SLOWDOWN,ALERT
	OverLimitActionList []string `gorm:"-" json:"over_limit_action_list,omitempty"`
	Status              string   `json:"status"`
	CreatedAt           int64    `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`
	UpdatedAt           int64    `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`
}

func (table *TrafficPolicy) TableName() string {
	return "traffic_policy"
}

func (table *TrafficPolicy) CountForName() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.Name != "" {
		tx = tx.Where("name = ?", table.Name)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}
