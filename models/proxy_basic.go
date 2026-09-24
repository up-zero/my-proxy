package models

// 代理类型
const (
	ProxyTypeTcp    = "TCP"
	ProxyTypeUdp    = "UDP"
	ProxyTypeHttp   = "HTTP"
	ProxyTypeSocks5 = "SOCKS5"
	ProxyTypeTcpUdp = "TCP_UDP" // TCP+UDP 双协议（仅用于创建时选择，落库时拆分为两条）
)

// 上游协议（仅 HTTP 固定转发使用）
const (
	ProxyUpstreamSchemeHttp  = "http"
	ProxyUpstreamSchemeHttps = "https"
)

// 代理状态
var (
	ProxyStateRunning = "RUNNING"
	ProxyStateStopped = "STOPPED"
)

type ProxyBasic struct {
	Uuid           string     `json:"uuid"`                                                      // 唯一标识
	Name           string     `json:"name"`                                                      // 代理名称
	TagUuidList    []string   `gorm:"-" json:"tag_uuid_list,omitempty"`                          // 标签唯一标识列表
	TagList        []TagBasic `gorm:"-" json:"tag_list,omitempty"`                               // 标签列表
	Type           string     `json:"type"`                                                      // 代理类型
	ListenAddress  string     `json:"listen_address"`                                            // 监听地址
	ListenPort     string     `json:"listen_port"`                                               // 监听端口
	TargetAddress  string     `json:"target_address"`                                            // 目标地址
	TargetPort     string     `json:"target_port"`                                               // 目标端口
	Socks5Username string     `json:"socks5_username"`                                           // SOCKS5 认证用户名
	Socks5Password string     `json:"socks5_password"`                                           // SOCKS5 认证密码
	HttpUsername   string     `json:"http_username"`                                             // HTTP 认证用户名
	HttpPassword   string     `json:"http_password"`                                             // HTTP 认证密码
	UpstreamScheme string     `json:"upstream_scheme"`                                           // 上游协议：http/https，仅 HTTP 固定转发使用
	State          string     `json:"state"`                                                     // 代理状态
	FailDetail     string     `json:"fail_detail"`                                               // 代理失败详情
	CreatedAt      int64      `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt      int64      `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *ProxyBasic) TableName() string {
	return "proxy_basic"
}

// IsHttpFixedForward HTTP 类型且配置了目标地址时为固定转发（反向代理），否则为动态代理
func (table *ProxyBasic) IsHttpFixedForward() bool {
	return table.Type == ProxyTypeHttp && table.TargetAddress != ""
}

// UpstreamSchemeOrDefault 获取上游协议，未显式配置时按目标端口推断
func (table *ProxyBasic) UpstreamSchemeOrDefault() string {
	if table.UpstreamScheme != "" {
		return table.UpstreamScheme
	}
	if table.TargetPort == "443" {
		return ProxyUpstreamSchemeHttps
	}
	return ProxyUpstreamSchemeHttp
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

// CountForPort 端口判重（同协议+同端口才视为重复，不同协议可共用端口）
func (table *ProxyBasic) CountForPort() (int64, error) {
	var cnt int64
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid != ?", table.Uuid)
	}
	if table.ListenAddress != "" {
		tx = tx.Where("listen_address = ?", table.ListenAddress)
	}
	if table.ListenPort != "" {
		tx = tx.Where("listen_port = ?", table.ListenPort)
	}
	if table.Type != "" {
		tx = tx.Where("type = ?", table.Type)
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
