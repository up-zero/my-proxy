package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Permission keys（菜单级别）
const (
	PermDashboardView = "dashboard.view"        // 仪表盘
	PermProxyView     = "proxy.view"            // 代理列表
	PermTagManage     = "tag.manage"            // 标签管理
	PermTrafficPolicy = "traffic_policy.manage" // 限速配额
	PermAlertView     = "alert.view"            // 告警通知
	PermAuditView     = "audit.view"            // 日志审计
	PermTerminalView  = "terminal.view"         // Web 终端
	PermUserManage    = "user.manage"           // 用户列表
	PermRoleManage    = "role.manage"           // 权限策略
)

// AllPermissions returns all available permission keys
func AllPermissions() []string {
	return []string{
		PermDashboardView,
		PermProxyView,
		PermTagManage,
		PermTrafficPolicy,
		PermAlertView,
		PermAuditView,
		PermTerminalView,
		PermUserManage,
		PermRoleManage,
	}
}

// BuiltIn role names
const (
	RoleNameAdmin = "admin"
	RoleNameOps   = "ops"
)

type RoleBasic struct {
	Uuid        string `json:"uuid" gorm:"primaryKey"`        // 唯一标识
	Name        string `json:"name" gorm:"uniqueIndex"`       // 角色名称
	Description string `json:"description"`                   // 角色描述
	BuiltIn     bool   `json:"built_in" gorm:"default:false"` // 是否内置角色
	Permissions string `json:"permissions" gorm:"type:text"`  // JSON array of permission keys
	CreatedAt   int64  `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`
}

func (r *RoleBasic) TableName() string {
	return "role_basic"
}

// GetPermissionList 解析权限列表
func (r *RoleBasic) GetPermissionList() []string {
	if r.Permissions == "" {
		return []string{}
	}
	var perms []string
	if err := json.Unmarshal([]byte(r.Permissions), &perms); err != nil {
		return []string{}
	}
	return perms
}

// SetPermissionList 设置权限列表
func (r *RoleBasic) SetPermissionList(perms []string) {
	data, _ := json.Marshal(perms)
	r.Permissions = string(data)
}

// HasPermission 检查是否有指定权限
func (r *RoleBasic) HasPermission(perm string) bool {
	for _, p := range r.GetPermissionList() {
		if p == perm {
			return true
		}
	}
	return false
}

// IsAdminRole 判断是否为管理员角色（拥有所有权限）
func (r *RoleBasic) IsAdminRole() bool {
	return r.Name == RoleNameAdmin && r.BuiltIn
}

// BeforeSave GORM hook to ensure permissions is valid JSON
func (r *RoleBasic) BeforeSave(tx *gorm.DB) error {
	if r.Permissions == "" {
		r.Permissions = "[]"
	}
	return nil
}

// AdminPermissions returns all permissions for admin role
func AdminPermissions() []string {
	return AllPermissions()
}

// OpsPermissions returns permissions for ops role
func OpsPermissions() []string {
	return []string{
		PermDashboardView,
		PermProxyView,
		PermTagManage,
		PermTrafficPolicy,
		PermAlertView,
		PermAuditView,
		PermTerminalView,
	}
}
