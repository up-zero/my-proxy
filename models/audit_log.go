package models

// 审计日志模块类型常量
const (
	AuditModuleAuth          = "AUTH"           // 用户登录
	AuditModuleProxy         = "PROXY"          // 代理管理
	AuditModuleTrafficPolicy = "TRAFFIC_POLICY" // 限速配额
)

// 审计日志操作类型常量
const (
	AuditActionLogin   = "LOGIN"   // 登录
	AuditActionCreate  = "CREATE"  // 新增
	AuditActionUpdate  = "UPDATE"  // 修改
	AuditActionDelete  = "DELETE"  // 删除
	AuditActionStart   = "START"   // 启动
	AuditActionStop    = "STOP"    // 停止
	AuditActionRestart = "RESTART" // 重启
	AuditActionEnable  = "ENABLE"  // 启用
	AuditActionDisable = "DISABLE" // 停用
	AuditActionImport  = "IMPORT"  // 导入
)

// AuditLog 审计日志记录
type AuditLog struct {
	Uuid       string `json:"uuid"`
	Username   string `json:"username"`    // 操作人用户名
	Module     string `json:"module"`      // 所属模块：AUTH / PROXY / TRAFFIC_POLICY
	Action     string `json:"action"`      // 操作类型：LOGIN / CREATE / UPDATE / DELETE / START / STOP / RESTART / ENABLE / DISABLE
	Target     string `json:"target"`      // 操作对象名称
	TargetUuid string `json:"target_uuid"` // 操作对象 UUID
	Detail     string `json:"detail"`      // 详情描述
	SourceIp   string `json:"source_ip"`   // 来源 IP
	CreatedAt  int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`
}

func (table *AuditLog) TableName() string {
	return "audit_log"
}
