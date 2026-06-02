package audit

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

const retention = 90 * 24 * time.Hour

func cleanupExpiredRecords() {
	expiredBefore := time.Now().Add(-retention).UnixMilli()
	if err := models.DB.Where("created_at < ?", expiredBefore).Delete(new(models.AuditLog)).Error; err != nil {
		logger.Error("[db] cleanup expired audit log records error.", zap.Error(err))
	}
}

// GetSourceIp 从 gin 上下文获取来源 IP
func GetSourceIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = c.RemoteIP()
	}
	return ip
}

// GetUsername 从 gin 上下文获取当前登录用户名
func GetUsername(c *gin.Context) string {
	claim, exists := c.Get("UserClaim")
	if !exists {
		return ""
	}
	if uc, ok := claim.(*util.UserClaim); ok {
		return uc.Username
	}
	return ""
}

// CreateRecord 创建审计日志记录
func CreateRecord(username, module, action, target, targetUuid, detail, sourceIp string) error {
	cleanupExpiredRecords()
	record := &models.AuditLog{
		Uuid:       idutil.UUIDGenerate(),
		Username:   strings.TrimSpace(username),
		Module:     strings.TrimSpace(module),
		Action:     strings.TrimSpace(action),
		Target:     strings.TrimSpace(target),
		TargetUuid: strings.TrimSpace(targetUuid),
		Detail:     strings.TrimSpace(detail),
		SourceIp:   strings.TrimSpace(sourceIp),
	}
	return models.DB.Create(record).Error
}

// LogWithContext 使用 gin.Context 自动提取用户名和来源 IP 并记录审计日志
func LogWithContext(c *gin.Context, module, action, target, targetUuid, detail string) {
	username := GetUsername(c)
	sourceIp := GetSourceIp(c)
	if err := CreateRecord(username, module, action, target, targetUuid, detail, sourceIp); err != nil {
		logger.Error("[audit] create audit log error.", zap.Error(err))
	}
}

// List 审计日志列表
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
	list := make([]*models.AuditLog, 0)
	retentionStart := time.Now().Add(-retention).UnixMilli()
	tx := models.DB.Model(new(models.AuditLog)).Where("created_at >= ?", retentionStart)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("username like ? OR target like ? OR detail like ?", like, like, like)
	}
	if module := strings.TrimSpace(in.Module); module != "" {
		tx = tx.Where("module = ?", strings.ToUpper(module))
	}
	if action := strings.TrimSpace(in.Action); action != "" {
		tx = tx.Where("action = ?", strings.ToUpper(action))
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		logger.Error("[db] get audit log count error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if err := tx.Order("created_at desc").Offset((in.Page - 1) * in.PerPage).Limit(in.PerPage).Find(&list).Error; err != nil {
		logger.Error("[db] get audit log list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOkWithList(c, list, count)
}
