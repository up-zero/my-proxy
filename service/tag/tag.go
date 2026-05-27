package tag

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

func savePreValid(tb *models.TagBasic) error {
	cnt, err := tb.CountForName()
	if err != nil {
		logger.Error("[db] get tag count for save error.", zap.Error(err))
		return err
	}
	if cnt > 0 {
		return errors.New(util.MsgErrNameExist)
	}
	return nil
}

// List 标签列表
func List(c *gin.Context, in *ListRequest) {
	list := make([]*models.TagBasic, 0)
	tx := models.DB.Model(new(models.TagBasic))
	if strings.TrimSpace(in.Name) != "" {
		tx = tx.Where("name like ?", "%"+strings.TrimSpace(in.Name)+"%")
	}
	if err := tx.Order("created_at desc").Find(&list).Error; err != nil {
		logger.Error("[db] get tag list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOkWithList(c, list)
}

// Create 创建标签
func Create(c *gin.Context, in *CreateRequest) {
	tb := &models.TagBasic{
		Uuid: idutil.UUIDGenerate(),
		Name: strings.TrimSpace(in.Name),
	}
	if err := savePreValid(tb); err != nil {
		util.ResponseError(c, err)
		return
	}
	if err := models.DB.Create(tb).Error; err != nil {
		logger.Error("[db] tag create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Update 编辑标签
func Update(c *gin.Context, in *UpdateRequest) {
	tb := &models.TagBasic{
		Uuid: in.Uuid,
		Name: strings.TrimSpace(in.Name),
	}
	if err := savePreValid(tb); err != nil {
		util.ResponseError(c, err)
		return
	}
	if err := models.DB.Model(new(models.TagBasic)).Where("uuid = ?", in.Uuid).Update("name", tb.Name).Error; err != nil {
		logger.Error("[db] tag update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Delete 删除标签
func Delete(c *gin.Context, in *DeleteRequest) {
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_uuid = ?", in.Uuid).Delete(new(models.ProxyTag)).Error; err != nil {
			return err
		}
		if err := tx.Where("uuid = ?", in.Uuid).Delete(new(models.TagBasic)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Error("[db] tag delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}
