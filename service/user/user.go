package user

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/audit"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Login 用户登录
func Login(c *gin.Context, in *LoginRequest) {
	// 用户鉴权
	uc := new(util.UserClaim)
	ub := &models.UserBasic{Username: in.Username, Password: in.Password}
	if err := ub.First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.ResponseMsg(c, util.CodeErr, util.MsgErrUsernameOrPassword)
			return
		}
		logger.Error("[DB] ERROR.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if err := convertutil.CopyProperties(ub, uc); err != nil {
		logger.Error("[gotool] copy properties error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	// 生成 token
	token, err := uc.GenerateToken(time.Now().Add(time.Hour * 24).Unix())
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}
	// 生成 refreshToken
	refreshToken, err := uc.GenerateToken(time.Now().Add(time.Hour * 24 * 2).Unix())
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}

	// 获取角色信息
	var roleName string
	var permissions []string
	if ub.Level == models.UserLevelRoot {
		// root 用户拥有所有权限
		roleName = models.RoleNameAdmin
		permissions = models.AdminPermissions()
	} else if ub.RoleID != "" {
		role := &models.RoleBasic{Uuid: ub.RoleID}
		if err := models.DB.First(role).Error; err == nil {
			roleName = role.Name
			permissions = role.GetPermissionList()
		}
	}

	// 记录登录审计日志
	_ = audit.CreateRecord(in.Username, models.AuditModuleAuth, models.AuditActionLogin, in.Username, ub.Uuid, "用户登录成功", audit.GetSourceIp(c))

	util.ResponseOkWithData(c, &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Username:     uc.Username,
		Level:        uc.Level,
		RoleID:       ub.RoleID,
		RoleName:     roleName,
		Permissions:  permissions,
	})
}

// RefreshToken 刷新 token
func RefreshToken(c *gin.Context, in *RefreshTokenRequest) {
	uc, err := util.AnalyzeToken(in.RefreshToken)
	if err != nil || uc.Username == "" {
		util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
		return
	}
	// 用户鉴权
	ub := &models.UserBasic{Username: uc.Username}
	if err = ub.First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.ResponseMsg(c, util.CodeErr, util.MsgErrAuth)
			return
		}
		logger.Error("[DB] ERROR.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	// 生成 token
	token, err := uc.GenerateToken(time.Now().Add(time.Hour * 24).Unix())
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}
	// 生成 refreshToken
	refreshToken, err := uc.GenerateToken(time.Now().Add(time.Hour * 24 * 2).Unix())
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}

	// 获取角色信息
	var roleName string
	var permissions []string
	if ub.Level == models.UserLevelRoot {
		roleName = models.RoleNameAdmin
		permissions = models.AdminPermissions()
	} else if ub.RoleID != "" {
		role := &models.RoleBasic{Uuid: ub.RoleID}
		if err := models.DB.First(role).Error; err == nil {
			roleName = role.Name
			permissions = role.GetPermissionList()
		}
	}

	util.ResponseOkWithData(c, &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Username:     uc.Username,
		Level:        uc.Level,
		RoleID:       ub.RoleID,
		RoleName:     roleName,
		Permissions:  permissions,
	})
}

// EditPassword 修改密码
func EditPassword(c *gin.Context, in *EditPasswordRequest) {
	uc := c.MustGet("UserClaim").(*util.UserClaim)
	// 判断旧密码是否一致
	if err := (&models.UserBasic{Username: uc.Username, Password: in.OldPassword}).First(); err != nil {
		logger.Error("[db] get user basic error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErr, util.MsgErrOldPasswordWrong)
		return
	}
	// 落库
	if err := models.DB.Model(new(models.UserBasic)).Where("username = ?", uc.Username).
		Update("password", in.NewPassword).Error; err != nil {
		logger.Error("[db] update user password error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}

	util.ResponseOk(c)
}

// List 用户列表
func List(c *gin.Context) {
	list := make([]*models.UserBasic, 0)
	if err := models.DB.Model(new(models.UserBasic)).Where("level != ?", models.UserLevelRoot).
		Find(&list).Error; err != nil {
		logger.Error("[db] get user list error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	util.ResponseOkWithList(c, list)
}

// Create 创建用户
func Create(c *gin.Context, in *CreateRequest) {
	ub := &models.UserBasic{
		Uuid:     idutil.UUIDGenerate(),
		Username: in.Username,
		Password: in.Password,
		Level:    models.UserLevelUser,
		RoleID:   in.RoleID,
	}
	// 用户名判重
	cnt, err := ub.CountForSave()
	if err != nil {
		logger.Error("[db] get user count for save error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if cnt > 0 {
		util.ResponseMsg(c, util.CodeErr, util.MsgErrNameExist)
		return
	}
	// 落库
	err = models.DB.Create(ub).Error
	if err != nil {
		logger.Error("[db] user create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}

	util.ResponseOk(c)
}

// Update 编辑用户
func Update(c *gin.Context, in *UpdateRequest) {
	ub := &models.UserBasic{
		Uuid:     in.Uuid,
		Username: in.Username,
	}
	// 用户名判重
	cnt, err := ub.CountForSave()
	if err != nil {
		logger.Error("[db] get user count for save error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if cnt > 0 {
		util.ResponseMsg(c, util.CodeErr, util.MsgErrNameExist)
		return
	}
	// 落库
	err = models.DB.Model(new(models.UserBasic)).Where("uuid = ?", in.Uuid).Updates(map[string]interface{}{
		"username": in.Username,
		"password": in.Password,
		"role_id":  in.RoleID,
	}).Error
	if err != nil {
		logger.Error("[db] user update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}

	util.ResponseOk(c)
}

// Delete 删除用户
func Delete(c *gin.Context, in *DeleteRequest) {
	// 删除数据
	err := models.DB.Where("uuid in ?", in.Uuid).
		Delete(new(models.UserBasic)).Error
	if err != nil {
		logger.Error("[db] user delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}
