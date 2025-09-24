package models

// 代理类型
const (
	ProxyTypeTcp  = "TCP"
	ProxyTypeUdp  = "UDP"
	ProxyTypeHttp = "HTTP"
)

// 代理状态
var (
	ProxyStateRunning = "RUNNING"
	ProxyStateStopped = "STOPPED"
)

type ProxyBasic struct {
	Uuid          string `json:"uuid"`                                                      // 唯一标识
	Name          string `json:"name"`                                                      // 代理名称
	Type          string `json:"type"`                                                      // 代理类型
	ListenPort    string `json:"listen_port"`                                               // 监听端口
	TargetAddress string `json:"target_address"`                                            // 目标地址
	TargetPort    string `json:"target_port"`                                               // 目标端口
	State         string `json:"state"`                                                     // 代理状态
	FailDetail    string `json:"fail_detail"`                                               // 代理失败详情
	CreatedAt     int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt     int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *ProxyBasic) TableName() string {
	return "proxy_basic"
}

// CountForName 保存时名称判重
func (table *ProxyBasic) CountForName() (int64, error) {
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

// CountForPort 端口判重
func (table *ProxyBasic) CountForPort() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.ListenPort != "" {
		tx = tx.Where("listen_port = ?", table.ListenPort)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}

// All 获取所有代理
func (table *ProxyBasic) All() ([]*ProxyBasic, error) {
	list := make([]*ProxyBasic, 0)
	return list, DB.Model(table).Find(&list).Error
}

// First 获取单个代理
func (table *ProxyBasic) First() error {
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid = ?", table.Uuid)
	}
	if table.Name != "" {
		tx = tx.Where("name = ?", table.Name)
	}
	return tx.First(table).Error
}
