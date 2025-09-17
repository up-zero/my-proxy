package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
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

	util.ResponseOkWithData(c, &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Username:     uc.Username,
		Level:        uc.Level,
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

	util.ResponseOkWithData(c, &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Username:     uc.Username,
		Level:        uc.Level,
	})
}
