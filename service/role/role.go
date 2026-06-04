package role

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// List 角色列表
func List(c *gin.Context, in *ListRequest) {
	list := make([]*models.RoleBasic, 0)
	tx := models.DB.Model(new(models.RoleBasic))
	if strings.TrimSpace(in.Name) != "" {
		tx = tx.Where("name like ?", "%"+strings.TrimSpace(in.Name)+"%")
	}
	if err := tx.Order("created_at asc").Find(&list).Error; err != nil {
		logger.Error("[db] get role list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOkWithList(c, list)
}

// Create 创建角色
func Create(c *gin.Context, in *CreateRequest) {
	r := &models.RoleBasic{
		Uuid:        idutil.UUIDGenerate(),
		Name:        strings.TrimSpace(in.Name),
		Description: in.Description,
		BuiltIn:     false,
	}
	r.SetPermissionList(in.Permissions)

	// 名称判重
	cnt, err := countByName(r.Name, "")
	if err != nil {
		logger.Error("[db] get role count error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if cnt > 0 {
		util.ResponseMsg(c, util.CodeErr, util.MsgErrNameExist)
		return
	}

	if err := models.DB.Create(r).Error; err != nil {
		logger.Error("[db] role create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Update 更新角色
func Update(c *gin.Context, in *UpdateRequest) {
	// 检查是否为内置角色
	existing := &models.RoleBasic{}
	if err := models.DB.Where("uuid = ?", in.Uuid).First(existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.ResponseMsg(c, util.CodeErr, "角色不存在")
			return
		}
		logger.Error("[db] get role error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if existing.BuiltIn && existing.Name == models.RoleNameAdmin {
		util.ResponseMsg(c, util.CodeErr, "管理员角色不可修改")
		return
	}

	// 名称判重（排除自身）
	cnt, err := countByName(strings.TrimSpace(in.Name), in.Uuid)
	if err != nil {
		logger.Error("[db] get role count error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if cnt > 0 {
		util.ResponseMsg(c, util.CodeErr, util.MsgErrNameExist)
		return
	}

	role := &models.RoleBasic{}
	role.SetPermissionList(in.Permissions)

	if err := models.DB.Model(new(models.RoleBasic)).Where("uuid = ?", in.Uuid).Updates(map[string]interface{}{
		"name":        strings.TrimSpace(in.Name),
		"description": in.Description,
		"permissions": role.Permissions,
	}).Error; err != nil {
		logger.Error("[db] role update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Delete 删除角色
func Delete(c *gin.Context, in *DeleteRequest) {
	// 检查是否有内置角色
	var builtinCount int64
	if err := models.DB.Model(new(models.RoleBasic)).Where("uuid in ? AND built_in = ?", in.Uuid, true).
		Count(&builtinCount).Error; err != nil {
		logger.Error("[db] check builtin role error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if builtinCount > 0 {
		util.ResponseMsg(c, util.CodeErr, "内置角色不可删除")
		return
	}

	// 检查是否有用户绑定该角色
	var userCount int64
	if err := models.DB.Model(new(models.UserBasic)).Where("role_id in ?", in.Uuid).
		Count(&userCount).Error; err != nil {
		logger.Error("[db] check role users error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if userCount > 0 {
		util.ResponseMsg(c, util.CodeErr, "该角色下存在用户，不可删除")
		return
	}

	if err := models.DB.Where("uuid in ?", in.Uuid).Delete(new(models.RoleBasic)).Error; err != nil {
		logger.Error("[db] role delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// countByName 按名称统计角色数量（排除指定UUID）
func countByName(name string, excludeUUID string) (int64, error) {
	var cnt int64
	tx := models.DB.Model(new(models.RoleBasic))
	if name != "" {
		tx = tx.Where("name = ?", name)
	}
	if excludeUUID != "" {
		tx = tx.Where("uuid != ?", excludeUUID)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}
