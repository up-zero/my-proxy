package alert

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

const SourceTrafficPolicy = "TRAFFIC_POLICY"
const retention = 30 * 24 * time.Hour

func cleanupExpiredRecords() {
	expiredBefore := time.Now().Add(-retention).UnixMilli()
	if err := models.DB.Where("created_at < ?", expiredBefore).Delete(new(models.AlertRecord)).Error; err != nil {
		logger.Error("[db] cleanup expired alert records error.", zap.Error(err))
	}
}

func CreateRecord(source, sourceUuid, level, title, content string) error {
	cleanupExpiredRecords()
	record := &models.AlertRecord{
		Uuid:       idutil.UUIDGenerate(),
		Source:     source,
		SourceUuid: sourceUuid,
		Level:      level,
		Title:      strings.TrimSpace(title),
		Content:    strings.TrimSpace(content),
	}
	return models.DB.Create(record).Error
}

func List(c *gin.Context, in *ListRequest) {
	cleanupExpiredRecords()
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PerPage <= 0 {
		in.PerPage = 20
	}
	if in.PerPage < 20 {
		in.PerPage = 20
	}
	if in.PerPage > 500 {
		in.PerPage = 500
	}
	list := make([]*models.AlertRecord, 0)
	retentionStart := time.Now().Add(-retention).UnixMilli()
	tx := models.DB.Model(new(models.AlertRecord)).Where("created_at >= ?", retentionStart)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("title like ? OR content like ? OR source like ?", like, like, like)
	}
	if level := strings.TrimSpace(in.Level); level != "" {
		tx = tx.Where("level = ?", level)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		logger.Error("[db] get alert count error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if err := tx.Order("created_at desc").Offset((in.Page - 1) * in.PerPage).Limit(in.PerPage).Find(&list).Error; err != nil {
		logger.Error("[db] get alert list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOkWithList(c, list, count)
}
