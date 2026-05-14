package group

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

func savePreValid(gb *models.GroupBasic) error {
	cnt, err := gb.CountForName()
	if err != nil {
		logger.Error("[db] get group count for save error.", zap.Error(err))
		return err
	}
	if cnt > 0 {
		return errors.New(util.MsgErrNameExist)
	}
	return nil
}

// List 分组列表
func List(c *gin.Context, in *ListRequest) {
	list := make([]*models.GroupBasic, 0)
	tx := models.DB.Model(new(models.GroupBasic))
	if strings.TrimSpace(in.Name) != "" {
		tx = tx.Where("name like ?", "%"+strings.TrimSpace(in.Name)+"%")
	}
	if err := tx.Order("created_at desc").Find(&list).Error; err != nil {
		logger.Error("[db] get group list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOkWithList(c, list)
}

// Create 创建分组
func Create(c *gin.Context, in *CreateRequest) {
	gb := &models.GroupBasic{
		Uuid: idutil.UUIDGenerate(),
		Name: strings.TrimSpace(in.Name),
	}
	if err := savePreValid(gb); err != nil {
		util.ResponseError(c, err)
		return
	}
	if err := models.DB.Create(gb).Error; err != nil {
		logger.Error("[db] group create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Update 编辑分组
func Update(c *gin.Context, in *UpdateRequest) {
	gb := &models.GroupBasic{
		Uuid: in.Uuid,
		Name: strings.TrimSpace(in.Name),
	}
	if err := savePreValid(gb); err != nil {
		util.ResponseError(c, err)
		return
	}
	if err := models.DB.Model(new(models.GroupBasic)).Where("uuid = ?", in.Uuid).Update("name", gb.Name).Error; err != nil {
		logger.Error("[db] group update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Delete 删除分组
func Delete(c *gin.Context, in *DeleteRequest) {
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(new(models.ProxyBasic)).Where("group_uuid = ?", in.Uuid).Update("group_uuid", "").Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("[db] group delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}
