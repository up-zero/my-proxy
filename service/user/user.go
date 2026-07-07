package user

import (
	"errors"
	"strings"
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
	// 读取 Token 有效期配置（天），0 表示永不过期
	tokenExpiryDays := models.GetConfigInt(util.ConfigKeyTokenExpiryDays, 1)

	// 生成 token
	token, err := uc.GenerateToken(getTokenExpireAt(tokenExpiryDays))
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}
	// 生成 refreshToken（有效期是 token 的 2 倍，永不过期则同样永不过期）
	refreshTokenExpiryDays := tokenExpiryDays
	if tokenExpiryDays > 0 {
		refreshTokenExpiryDays = tokenExpiryDays * 2
	}
	refreshToken, err := uc.GenerateToken(getTokenExpireAt(refreshTokenExpiryDays))
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
	// 读取 Token 有效期配置（天），0 表示永不过期
	tokenExpiryDays := models.GetConfigInt(util.ConfigKeyTokenExpiryDays, 1)

	// 生成 token
	token, err := uc.GenerateToken(getTokenExpireAt(tokenExpiryDays))
	if err != nil {
		util.ResponseMsg(c, util.CodeErr, err.Error())
		return
	}
	// 生成 refreshToken（有效期是 token 的 2 倍，永不过期则同样永不过期）
	refreshTokenExpiryDays := tokenExpiryDays
	if tokenExpiryDays > 0 {
		refreshTokenExpiryDays = tokenExpiryDays * 2
	}
	refreshToken, err := uc.GenerateToken(getTokenExpireAt(refreshTokenExpiryDays))
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

// getTokenExpireAt 根据配置的天数返回 Unix 时间戳。
// days<0（即 -1）表示永不过期，返回 0（JWT 不设置 ExpiresAt）。
func getTokenExpireAt(days int) int64 {
	if days < 0 {
		return 0
	}
	return time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()
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
func List(c *gin.Context, in *ListRequest) {
	list := make([]*models.UserBasic, 0)
	tx := models.DB.Model(new(models.UserBasic)).Where("level != ?", models.UserLevelRoot)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("username like ?", like)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		logger.Error("[db] get user count error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if err := tx.Order("created_at desc").Offset((in.Page - 1) * in.PerPage).Limit(in.PerPage).Find(&list).Error; err != nil {
		logger.Error("[db] get user list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}

	util.ResponseOkWithList(c, list, count)
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
