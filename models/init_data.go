package models

import (
	"github.com/up-zero/gotool/randomutil"
	"github.com/up-zero/my-proxy/logger"
	"go.uber.org/zap"
)

// initData 初始化数据
func initData() {
	// 初始化角色数据
	initRoleData()
	// 初始化用户
	initUserData()
}

func initRoleData() {
	roles := []RoleBasic{
		{
			Uuid:        "role-admin",
			Name:        RoleNameAdmin,
			Description: "超级管理员，拥有所有权限",
			BuiltIn:     true,
		},
		{
			Uuid:        "role-ops",
			Name:        RoleNameOps,
			Description: "运维人员，拥有代理管理和运维中心权限",
			BuiltIn:     true,
		},
	}

	for i := range roles {
		roles[i].SetPermissionList(getDefaultPermissions(roles[i].Name))
		if err := DB.FirstOrCreate(&roles[i], &RoleBasic{Uuid: roles[i].Uuid}).Error; err != nil {
			panic(err)
		}
	}

	// 更新超级管理员权限
	syncAdminPermissions()
}

// syncAdminPermissions 更新超级管理员权限
func syncAdminPermissions() {
	adminRole := &RoleBasic{Uuid: "role-admin"}
	if err := DB.First(adminRole).Error; err != nil {
		return
	}
	// 如果超级管理员权限不完整，自动更新
	currentPerms := adminRole.GetPermissionList()
	allPerms := AdminPermissions()
	if len(currentPerms) != len(allPerms) {
		adminRole.SetPermissionList(allPerms)
		if err := DB.Save(adminRole).Error; err != nil {
			logger.Error("Failed to sync admin permissions", zap.Error(err))
		}
	}
}

func getDefaultPermissions(roleName string) []string {
	switch roleName {
	case RoleNameAdmin:
		return AdminPermissions()
	case RoleNameOps:
		return OpsPermissions()
	default:
		return []string{}
	}
}

func initUserData() {
	// admin
	ub := &UserBasic{
		Username: "admin",
		Password: randomutil.Alphanumeric(8),
		Level:    "root",
		RoleID:   "role-admin",
	}
	if err := DB.FirstOrCreate(ub, &UserBasic{Username: ub.Username}).Error; err != nil {
		panic(err)
	}
}
