package trafficpolicy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/alert"
	"github.com/up-zero/my-proxy/service/audit"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

const (
	ScopeAll   = "ALL"
	ScopeTag   = "TAG"
	ScopeProxy = "PROXY"
)

func normalizeActionList(actions []string) []string {
	res := make([]string, 0, len(actions))
	seen := map[string]struct{}{}
	for _, action := range actions {
		action = strings.ToUpper(strings.TrimSpace(action))
		if action != models.OverLimitActionSlowdown && action != models.OverLimitActionAlert {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		res = append(res, action)
	}
	return res
}

func splitActions(action string) []string {
	if strings.TrimSpace(action) == "" {
		return nil
	}
	return normalizeActionList(strings.Split(action, ","))
}

func hasAction(action string, target string) bool {
	for _, item := range splitActions(action) {
		if item == target {
			return true
		}
	}
	return false
}

func normalizePolicy(policy *models.TrafficPolicy) {
	policy.Name = strings.TrimSpace(policy.Name)
	policy.ScopeType = strings.ToUpper(strings.TrimSpace(policy.ScopeType))
	if policy.ScopeType == "" {
		policy.ScopeType = ScopeAll
	}
	policy.ScopeValue = strings.TrimSpace(policy.ScopeValue)
	policy.OutboundLimit = strings.TrimSpace(policy.OutboundLimit)
	policy.MaxConnections = strings.TrimSpace(policy.MaxConnections)
	policy.PeriodQuota = strings.TrimSpace(policy.PeriodQuota)
	policy.OverLimitActionList = normalizeActionList(policy.OverLimitActionList)
	policy.OverLimitAction = strings.Join(policy.OverLimitActionList, ",")
	policy.Status = strings.ToUpper(strings.TrimSpace(policy.Status))
	if policy.Status == "" {
		policy.Status = models.TrafficPolicyStatusEnabled
	}
}

func validatePolicy(policy *models.TrafficPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("name is required")
	}
	if policy.ScopeType != ScopeAll && policy.ScopeType != ScopeTag && policy.ScopeType != ScopeProxy {
		return fmt.Errorf("scope_type is invalid")
	}
	if policy.ScopeType != ScopeAll && policy.ScopeValue == "" {
		return fmt.Errorf("scope_value is required")
	}
	if policy.OutboundLimit == "" && policy.MaxConnections == "" && policy.PeriodQuota == "" {
		return fmt.Errorf("outbound_limit, max_connections or period_quota is required")
	}
	if len(policy.OverLimitActionList) == 0 {
		return fmt.Errorf("over_limit_action_list is required")
	}
	if policy.Status != models.TrafficPolicyStatusEnabled && policy.Status != models.TrafficPolicyStatusDisabled {
		return fmt.Errorf("status is invalid")
	}
	cnt, err := policy.CountForName()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf(util.MsgErrNameExist)
	}
	return nil
}

func buildPolicy(in *SaveRequest) *models.TrafficPolicy {
	policy := &models.TrafficPolicy{
		Uuid:                strings.TrimSpace(in.Uuid),
		Name:                in.Name,
		ScopeType:           in.ScopeType,
		ScopeValue:          in.ScopeValue,
		OutboundLimit:       in.OutboundLimit,
		MaxConnections:      in.MaxConnections,
		PeriodQuota:         in.PeriodQuota,
		OverLimitActionList: in.OverLimitActionList,
	}
	normalizePolicy(policy)
	return policy
}

func populateScopeName(list []*models.TrafficPolicy) {
	if len(list) == 0 {
		return
	}
	proxyUuids := make([]string, 0)
	tagUuids := make([]string, 0)
	for _, item := range list {
		switch item.ScopeType {
		case ScopeProxy:
			proxyUuids = append(proxyUuids, item.ScopeValue)
		case ScopeTag:
			tagUuids = append(tagUuids, item.ScopeValue)
		}
	}
	proxyMap := map[string]string{}
	if len(proxyUuids) > 0 {
		var proxies []models.ProxyBasic
		if err := models.DB.Model(new(models.ProxyBasic)).Where("uuid in ?", proxyUuids).Find(&proxies).Error; err == nil {
			for _, item := range proxies {
				proxyMap[item.Uuid] = item.Name
			}
		}
	}
	tagMap := map[string]string{}
	if len(tagUuids) > 0 {
		var tags []models.TagBasic
		if err := models.DB.Model(new(models.TagBasic)).Where("uuid in ?", tagUuids).Find(&tags).Error; err == nil {
			for _, item := range tags {
				tagMap[item.Uuid] = item.Name
			}
		}
	}
	for _, item := range list {
		item.OverLimitActionList = splitActions(item.OverLimitAction)
		switch item.ScopeType {
		case ScopeAll:
			item.ScopeName = "全部代理"
		case ScopeProxy:
			item.ScopeName = proxyMap[item.ScopeValue]
		case ScopeTag:
			item.ScopeName = tagMap[item.ScopeValue]
		}
		if item.ScopeName == "" {
			item.ScopeName = item.ScopeValue
		}
	}
}

func parseQuotaBytes(value string) (float64, bool) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return 0, false
	}
	re := regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*(B|KB|MB|GB|TB)?`)
	matches := re.FindStringSubmatch(value)
	if len(matches) == 0 {
		return 0, false
	}
	number, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, false
	}
	unit := "B"
	if len(matches) > 2 && matches[2] != "" {
		unit = matches[2]
	}
	multiplier := 1.0
	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	return number * multiplier, true
}

func createAlertIfTriggered(policy *models.TrafficPolicy, reason string) error {
	if policy.Status != models.TrafficPolicyStatusEnabled || !hasAction(policy.OverLimitAction, models.OverLimitActionAlert) {
		return nil
	}
	triggered := strings.TrimSpace(reason) != ""
	if !triggered && policy.PeriodQuota != "" && policy.QuotaUsed != "" {
		quota, okQuota := parseQuotaBytes(policy.PeriodQuota)
		used, okUsed := parseQuotaBytes(policy.QuotaUsed)
		triggered = okQuota && okUsed && quota > 0 && used >= quota
	}
	if !triggered {
		return nil
	}
	if strings.TrimSpace(reason) == "" {
		reason = fmt.Sprintf("配额已用 %s，达到或超过周期配额 %s", policy.QuotaUsed, policy.PeriodQuota)
	}
	return alert.CreateRecord(alert.SourceTrafficPolicy, policy.Uuid, models.AlertLevelWarning, "限速配额规则触发告警", fmt.Sprintf("策略【%s】已触发：%s", policy.Name, reason))
}

func List(c *gin.Context, in *ListRequest) {
	list := make([]*models.TrafficPolicy, 0)
	tx := models.DB.Model(new(models.TrafficPolicy))
	if name := strings.TrimSpace(in.Name); name != "" {
		tx = tx.Where("name like ?", "%"+name+"%")
	}
	if scopeType := strings.TrimSpace(in.ScopeType); scopeType != "" {
		tx = tx.Where("scope_type = ?", strings.ToUpper(scopeType))
	}
	if status := strings.TrimSpace(in.Status); status != "" {
		tx = tx.Where("status = ?", strings.ToUpper(status))
	}
	if err := tx.Order("created_at desc").Find(&list).Error; err != nil {
		logger.Error("[db] get traffic policy list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	populateScopeName(list)
	quotaUsedMap := RuntimeQuotaUsedMap()
	for _, item := range list {
		if quotaUsed, ok := quotaUsedMap[item.Uuid]; ok {
			item.QuotaUsed = quotaUsed
		}
	}
	util.ResponseOkWithList(c, list)
}

func Create(c *gin.Context, in *SaveRequest) {
	policy := buildPolicy(in)
	policy.Uuid = idutil.UUIDGenerate()
	policy.Status = models.TrafficPolicyStatusEnabled
	if err := validatePolicy(policy); err != nil {
		util.ResponseError(c, err)
		return
	}
	if err := models.DB.Create(policy).Error; err != nil {
		logger.Error("[db] traffic policy create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	RefreshRuntime()
	if err := createAlertIfTriggered(policy, ""); err != nil {
		logger.Error("[db] traffic policy alert create error.", zap.Error(err))
	}
	audit.LogWithContext(c, models.AuditModuleTrafficPolicy, models.AuditActionCreate, policy.Name, policy.Uuid, fmt.Sprintf("新增限速策略：%s", policy.Name))
	util.ResponseOk(c)
}

func Update(c *gin.Context, in *SaveRequest) {
	policy := buildPolicy(in)
	if policy.Uuid == "" {
		util.ResponseMsg(c, util.CodeErrParam, util.MsgErrParam)
		return
	}
	oldPolicy := &models.TrafficPolicy{}
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("uuid = ?", policy.Uuid).First(oldPolicy).Error; err != nil {
		logger.Error("[db] traffic policy get for update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDataNotExist, util.MsgErrDataNotExist)
		return
	}
	policy.Status = oldPolicy.Status
	policy.QuotaUsed = oldPolicy.QuotaUsed
	if err := validatePolicy(policy); err != nil {
		util.ResponseError(c, err)
		return
	}
	updates := map[string]any{
		"name":              policy.Name,
		"scope_type":        policy.ScopeType,
		"scope_value":       policy.ScopeValue,
		"outbound_limit":    policy.OutboundLimit,
		"max_connections":   policy.MaxConnections,
		"period_quota":      policy.PeriodQuota,
		"over_limit_action": policy.OverLimitAction,
	}
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("uuid = ?", policy.Uuid).Updates(updates).Error; err != nil {
		logger.Error("[db] traffic policy update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	RefreshRuntime()
	if err := createAlertIfTriggered(policy, ""); err != nil {
		logger.Error("[db] traffic policy alert create error.", zap.Error(err))
	}
	audit.LogWithContext(c, models.AuditModuleTrafficPolicy, models.AuditActionUpdate, policy.Name, policy.Uuid, fmt.Sprintf("修改限速策略：%s", policy.Name))
	util.ResponseOk(c)
}

func Enable(c *gin.Context, in *UuidRequest) {
	updateStatus(c, in.Uuid, models.TrafficPolicyStatusEnabled)
}

func Disable(c *gin.Context, in *UuidRequest) {
	updateStatus(c, in.Uuid, models.TrafficPolicyStatusDisabled)
}

func updateStatus(c *gin.Context, uuid string, status string) {
	// 查询策略名称用于审计日志
	policy := &models.TrafficPolicy{Uuid: uuid}
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("uuid = ?", uuid).First(policy).Error; err != nil {
		logger.Error("[db] traffic policy get for status update error.", zap.Error(err))
	}
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("uuid = ?", uuid).Update("status", status).Error; err != nil {
		logger.Error("[db] traffic policy status update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	RefreshRuntime()
	if status == models.TrafficPolicyStatusEnabled {
		if policy.Name != "" {
			_ = createAlertIfTriggered(policy, "")
		}
		audit.LogWithContext(c, models.AuditModuleTrafficPolicy, models.AuditActionEnable, policy.Name, uuid, fmt.Sprintf("启用限速策略：%s", policy.Name))
	} else {
		audit.LogWithContext(c, models.AuditModuleTrafficPolicy, models.AuditActionDisable, policy.Name, uuid, fmt.Sprintf("停用限速策略：%s", policy.Name))
	}
	util.ResponseOk(c)
}

func Delete(c *gin.Context, in *UuidRequest) {
	// 查询策略名称用于审计日志
	policy := &models.TrafficPolicy{Uuid: in.Uuid}
	if err := models.DB.Model(new(models.TrafficPolicy)).Where("uuid = ?", in.Uuid).First(policy).Error; err != nil {
		logger.Error("[db] traffic policy get for delete error.", zap.Error(err))
	}
	if err := models.DB.Where("uuid = ?", in.Uuid).Delete(new(models.TrafficPolicy)).Error; err != nil {
		logger.Error("[db] traffic policy delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	RefreshRuntime()
	audit.LogWithContext(c, models.AuditModuleTrafficPolicy, models.AuditActionDelete, policy.Name, in.Uuid, fmt.Sprintf("删除限速策略：%s", policy.Name))
	util.ResponseOk(c)
}
