package models

import (
	"github.com/up-zero/gotool/randomutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// initData 初始化数据
func initData() {
	// 初始化 JWT 签名密钥
	initJwtSecret()
	// 初始化角色数据
	initRoleData()
	// 初始化用户
	initUserData()
	// 初始化本地节点
	initLocalNode()
}

// initJwtSecret 初始化 JWT 签名密钥
func initJwtSecret() {
	row := &ConfigBasic{Key: util.ConfigKeyJwtSecret}
	if err := DB.Where("key = ?", util.ConfigKeyJwtSecret).First(row).Error; err == nil && row.Value != "" {
		// 已有密钥，直接使用
		util.JwtKey = row.Value
		return
	}

	secret := randomutil.Alphanumeric(32)

	// 持久化到数据库
	row.Key = util.ConfigKeyJwtSecret
	row.Value = secret
	if err := DB.Create(row).Error; err != nil {
		panic("failed to save JWT secret: " + err.Error())
	}

	util.JwtKey = secret
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

func initLocalNode() {
	localNode := &NodeBasic{
		Uuid:      "node-local",
		Name:      "Local",
		Address:   "",
		SecretKey: "",
		Enabled:   true,
		IsLocal:   true,
	}
	if err := DB.FirstOrCreate(localNode, &NodeBasic{Uuid: "node-local"}).Error; err != nil {
		logger.Error("[init] create local node error.", zap.Error(err))
	}
}
