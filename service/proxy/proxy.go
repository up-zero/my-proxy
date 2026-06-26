package proxy

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/gotool/sliceutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/audit"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func loadProxyBasicMap() (map[string]*models.ProxyBasic, error) {
	list := make([]*models.ProxyBasic, 0)
	if err := models.DB.Model(new(models.ProxyBasic)).Find(&list).Error; err != nil {
		logger.Error("[db] get proxy list error.", zap.Error(err))
		return nil, err
	}
	proxyMap := make(map[string]*models.ProxyBasic, len(list))
	for _, item := range list {
		proxyMap[item.Uuid] = item
	}
	return proxyMap, nil
}

func normalizeTagUuidList(tagUuidList []string) []string {
	res := make([]string, 0, len(tagUuidList))
	seen := make(map[string]struct{}, len(tagUuidList))
	for _, item := range tagUuidList {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		res = append(res, item)
	}
	return res
}

func applyProxyTagList(pb *models.ProxyBasic, tagList []models.TagBasic) {
	pb.TagList = append([]models.TagBasic(nil), tagList...)
	pb.TagUuidList = make([]string, 0, len(tagList))
	for _, tag := range tagList {
		pb.TagUuidList = append(pb.TagUuidList, tag.Uuid)
	}
}

func applyTaskTagList(task *serve.ProxyTask, tagList []models.TagBasic) {
	task.TagList = append([]models.TagBasic(nil), tagList...)
	task.TagUuidList = make([]string, 0, len(tagList))
	for _, tag := range tagList {
		task.TagUuidList = append(task.TagUuidList, tag.Uuid)
	}
}

func populateProxyTags(list []*models.ProxyBasic) error {
	if len(list) == 0 {
		return nil
	}
	proxyUuids := make([]string, 0, len(list))
	for _, item := range list {
		proxyUuids = append(proxyUuids, item.Uuid)
	}
	proxyTagMap, err := models.LoadProxyTagListMap(proxyUuids)
	if err != nil {
		logger.Error("[db] get proxy tag relations error.", zap.Error(err))
		return err
	}
	for _, item := range list {
		applyProxyTagList(item, proxyTagMap[item.Uuid])
	}
	return nil
}

func populateTaskTags(list []*serve.ProxyTask) error {
	if len(list) == 0 {
		return nil
	}
	proxyUuids := make([]string, 0, len(list))
	for _, item := range list {
		proxyUuids = append(proxyUuids, item.Uuid)
	}
	proxyTagMap, err := models.LoadProxyTagListMap(proxyUuids)
	if err != nil {
		logger.Error("[db] get proxy tag relations error.", zap.Error(err))
		return err
	}
	for _, item := range list {
		applyTaskTagList(item, proxyTagMap[item.Uuid])
	}
	return nil
}

func hasAnyMatchedTag(proxyTagUuidList []string, selectedTagUuidList []string) bool {
	if len(selectedTagUuidList) == 0 {
		return true
	}
	selectedMap := make(map[string]struct{}, len(selectedTagUuidList))
	for _, item := range selectedTagUuidList {
		selectedMap[item] = struct{}{}
	}
	for _, item := range proxyTagUuidList {
		if _, ok := selectedMap[item]; ok {
			return true
		}
	}
	return false
}

func normalizeProxyBasic(pb *models.ProxyBasic) {
	pb.Name = strings.TrimSpace(pb.Name)
	if len(pb.TagUuidList) == 0 && len(pb.TagList) > 0 {
		for _, tag := range pb.TagList {
			pb.TagUuidList = append(pb.TagUuidList, tag.Uuid)
		}
	}
	pb.TagUuidList = normalizeTagUuidList(pb.TagUuidList)
	pb.Type = strings.ToUpper(strings.TrimSpace(pb.Type))
	pb.ListenAddress = strings.TrimSpace(pb.ListenAddress)
	pb.ListenPort = strings.TrimSpace(pb.ListenPort)
	pb.TargetAddress = strings.TrimSpace(pb.TargetAddress)
	pb.TargetPort = strings.TrimSpace(pb.TargetPort)
	pb.Socks5Username = strings.TrimSpace(pb.Socks5Username)
	pb.Socks5Password = strings.TrimSpace(pb.Socks5Password)

	if pb.Type == models.ProxyTypeSocks5 {
		pb.TargetAddress = ""
		pb.TargetPort = ""
	} else {
		// 非 SOCKS5 类型清空认证字段
		pb.Socks5Username = ""
		pb.Socks5Password = ""
	}
}

func isSupportedProxyType(proxyType string) bool {
	switch proxyType {
	case models.ProxyTypeTcp, models.ProxyTypeUdp, models.ProxyTypeHttp, models.ProxyTypeSocks5:
		return true
	default:
		return false
	}
}

func validateProxyBasicFields(pb *models.ProxyBasic) error {
	if pb.Name == "" {
		return fmt.Errorf("name is required")
	}
	if pb.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !isSupportedProxyType(pb.Type) {
		return fmt.Errorf("proxy type(%s) not support", pb.Type)
	}
	if pb.ListenPort == "" {
		return fmt.Errorf("listen_port is required")
	}
	if pb.Type != models.ProxyTypeSocks5 {
		if pb.TargetAddress == "" {
			return fmt.Errorf("target_address is required")
		}
		if pb.TargetPort == "" {
			return fmt.Errorf("target_port is required")
		}
	}
	return nil
}

// Status 获取代理状态
func Status(c *gin.Context, in *StatusRequest) {
	task := serve.ProxyTask{}
	if in.Uuid != "" {
		task.Uuid = in.Uuid
	}
	tasks, err := task.Status()
	if err != nil {
		logger.Error("[sys] proxy task status error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	proxyMap, err := loadProxyBasicMap()
	if err != nil {
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if err := populateTaskTags(tasks); err != nil {
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	for _, item := range tasks {
		if pb, ok := proxyMap[item.Uuid]; ok {
			item.Name = pb.Name
		}
	}
	if in.Name != "" {
		tasks = sliceutil.Filter(tasks, func(pt *serve.ProxyTask) bool {
			return strings.Contains(strings.ToLower(pt.Name), strings.ToLower(in.Name))
		})
	}
	if len(in.TagUuidList) > 0 {
		tasks = sliceutil.Filter(tasks, func(pt *serve.ProxyTask) bool {
			return hasAnyMatchedTag(pt.TagUuidList, in.TagUuidList)
		})
	}
	// 排序
	if in.SortField != "" && in.SortOrder != "" {
		sortTasks(tasks, in.SortField, in.SortOrder)
	}
	util.ResponseOkWithList(c, tasks)
}

// sortTasks 对代理任务列表进行排序
func sortTasks(tasks []*serve.ProxyTask, field string, order string) {
	less := func(i, j int) bool {
		var vi, vj string
		switch field {
		case "name":
			vi, vj = tasks[i].Name, tasks[j].Name
		case "type":
			vi, vj = tasks[i].Type, tasks[j].Type
		case "listen_address":
			vi, vj = tasks[i].ListenAddress, tasks[j].ListenAddress
		case "listen_port":
			pi, _ := strconv.Atoi(tasks[i].ListenPort)
			pj, _ := strconv.Atoi(tasks[j].ListenPort)
			if order == "descend" {
				return pi > pj
			}
			return pi < pj
		case "target_address":
			vi, vj = tasks[i].TargetAddress, tasks[j].TargetAddress
		case "target_port":
			pi, _ := strconv.Atoi(tasks[i].TargetPort)
			pj, _ := strconv.Atoi(tasks[j].TargetPort)
			if order == "descend" {
				return pi > pj
			}
			return pi < pj
		default:
			return false
		}
		if order == "descend" {
			return vi > vj
		}
		return vi < vj
	}
	sort.SliceStable(tasks, less)
}

// savePreValid 判断代理信息是否有效
func savePreValid(pb *models.ProxyBasic) error {
	normalizeProxyBasic(pb)
	if err := validateProxyBasicFields(pb); err != nil {
		return err
	}
	if len(pb.TagUuidList) > 0 {
		tagMap, err := models.LoadTagMap()
		if err != nil {
			logger.Error("[db] get tag map error.", zap.Error(err))
			return err
		}
		for _, tagUuid := range pb.TagUuidList {
			if _, ok := tagMap[tagUuid]; !ok {
				return fmt.Errorf("tag(%s) does not exist", tagUuid)
			}
		}
	}
	// 名称判重
	count, err := pb.CountForName()
	if err != nil {
		logger.Error("[sys] proxy basic count for name error.", zap.Error(err))
		return err
	}
	if count > 0 {
		return fmt.Errorf("name(%s) already exists", pb.Name)
	}
	// 端口判重
	count, err = pb.CountForPort()
	if err != nil {
		logger.Error("[sys] proxy basic count for port error.", zap.Error(err))
		return err
	}
	if count > 0 {
		return fmt.Errorf("listen_port(%s) already exists", pb.ListenPort)
	}
	return nil
}

// create 创建代理
func create(pb *models.ProxyBasic) error {
	// 代理信息校验
	if err := savePreValid(pb); err != nil {
		return err
	}

	// 启动代理任务
	task := serve.ProxyTask{
		ProxyBasic: *pb,
	}
	if err := task.Start(); err != nil {
		logger.Error("[sys] proxy task start error.", zap.Error(err))
		return err
	}

	// 落库
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pb).Error; err != nil {
			return err
		}
		return models.ReplaceProxyTags(tx, pb.Uuid, pb.TagUuidList)
	}); err != nil {
		task.Stop()
		task.Remove()
		logger.Error("[sys] proxy basic create error.", zap.Error(err))
		return err
	}

	return nil
}

// Create 创建代理
func Create(c *gin.Context, in *CreateRequest) {
	pb := new(models.ProxyBasic)
	if err := convertutil.CopyProperties(in, pb); err != nil {
		logger.Error("[gotool] copy properties error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	pb.Uuid = idutil.UUIDGenerate()
	pb.State = models.ProxyStateStopped

	// 创建代理
	if err := create(pb); err != nil {
		logger.Error("[sys] proxy basic create error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionCreate, pb.Name, pb.Uuid, fmt.Sprintf("新增代理：%s，类型：%s，监听端口：%s", pb.Name, pb.Type, pb.ListenPort))

	util.ResponseOk(c)
}

// Edit 编辑代理
func Edit(c *gin.Context, in *EditRequest) {
	pb := new(models.ProxyBasic)
	if err := convertutil.CopyProperties(in, pb); err != nil {
		logger.Error("[gotool] copy properties error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	pb.State = models.ProxyStateStopped

	// 代理信息校验
	if err := savePreValid(pb); err != nil {
		logger.Error("[sys] proxy basic save pre valid error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	// 重启代理任务
	task := serve.ProxyTask{
		ProxyBasic: *pb,
	}
	if err := task.Restart(); err != nil {
		logger.Error("[sys] proxy task restart error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	// 落库
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(new(models.ProxyBasic)).Where("uuid = ?", pb.Uuid).Updates(map[string]any{
			"name":            pb.Name,
			"type":            pb.Type,
			"listen_address":  pb.ListenAddress,
			"listen_port":     pb.ListenPort,
			"target_address":  pb.TargetAddress,
			"target_port":     pb.TargetPort,
			"socks5_username": pb.Socks5Username,
			"socks5_password": pb.Socks5Password,
			"state":           pb.State,
		}).Error; err != nil {
			return err
		}
		return models.ReplaceProxyTags(tx, pb.Uuid, pb.TagUuidList)
	}); err != nil {
		logger.Error("[sys] proxy basic updates error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionUpdate, pb.Name, pb.Uuid, fmt.Sprintf("修改代理：%s", pb.Name))

	util.ResponseOk(c)
}

// Export 导出代理
func Export(c *gin.Context, in *ExportRequest) {
	list := make([]*models.ProxyBasic, 0)
	if err := models.DB.Model(new(models.ProxyBasic)).Where("uuid IN ?", in.Uuid).Find(&list).Error; err != nil {
		logger.Error("[db] proxy basic find error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	if err := populateProxyTags(list); err != nil {
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}

	b, err := json.MarshalIndent(list, "", "\t")
	if err != nil {
		logger.Error("[sys] json marshal indent error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	util.ResponseFile(c, b, "proxy_list.json")
}

// Import 导入代理
func Import(c *gin.Context) {
	b, err := util.FormFileReadAll(c, "file")
	if err != nil {
		logger.Error("[util] form file read all error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	list := make([]*models.ProxyBasic, 0)
	if err := json.Unmarshal(b, &list); err != nil {
		logger.Error("[sys] json unmarshal error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	for _, pb := range list {
		// 创建代理
		pb.Uuid = idutil.UUIDGenerate()
		pb.State = models.ProxyStateStopped
		if err := create(pb); err != nil {
			logger.Error("[sys] proxy basic create error.", zap.Error(err))
			util.ResponseError(c, err)
			return
		}
		audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionImport, pb.Name, pb.Uuid, fmt.Sprintf("导入代理：%s", pb.Name))
	}

	util.ResponseOk(c)
}

// Delete 删除代理
func Delete(c *gin.Context, in *DeleteRequest) {
	// 判断代理是否存在
	pb := models.ProxyBasic{Name: in.Name}
	if err := pb.First(); err != nil {
		logger.Error("[db] proxy basic first error.", zap.Error(err))
		util.ResponseError(c, util.ErrProxyNotExist)
		return
	}

	// 停止代理
	task := serve.ProxyTask{
		ProxyBasic: pb,
	}
	if err := task.Stop(); err != nil {
		logger.Error("[sys] proxy task stop error.", zap.Error(err))
	}

	// 移除代理
	task.Remove()
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("proxy_uuid = ?", pb.Uuid).Delete(new(models.ProxyTag)).Error; err != nil {
			return err
		}
		return tx.Delete(new(models.ProxyBasic), "uuid = ?", pb.Uuid).Error
	}); err != nil {
		logger.Error("[sys] proxy basic delete error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionDelete, pb.Name, pb.Uuid, fmt.Sprintf("删除代理：%s", pb.Name))

	util.ResponseOk(c)
}

// BatchDelete 批量删除
func BatchDelete(c *gin.Context, in *BatchDeleteRequest) {
	// 批量停止
	for _, uuid := range in.Uuid {
		task := serve.ProxyTask{
			ProxyBasic: models.ProxyBasic{Uuid: uuid},
		}
		if err := task.Stop(); err != nil {
			logger.Error("[sys] proxy task stop error.", zap.Error(err))
		}
		task.Remove()
	}
	// 移除代理
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("proxy_uuid IN ?", in.Uuid).Delete(new(models.ProxyTag)).Error; err != nil {
			return err
		}
		return tx.Delete(new(models.ProxyBasic), "uuid in ?", in.Uuid).Error
	}); err != nil {
		logger.Error("[sys] proxy basic delete error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionDelete, "", "", fmt.Sprintf("批量删除代理，数量：%d", len(in.Uuid)))
	util.ResponseOk(c)
}

// Start 启动代理
func Start(c *gin.Context, in *StartRequest) {
	// 判断代理是否存在
	pb := models.ProxyBasic{Name: in.Name}
	if err := pb.First(); err != nil {
		logger.Error("[db] proxy basic first error.", zap.Error(err))
		util.ResponseError(c, util.ErrProxyNotExist)
		return
	}

	// 启动代理
	task := serve.ProxyTask{
		ProxyBasic: pb,
	}
	if err := task.Start(); err != nil {
		logger.Error("[sys] proxy task start error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionStart, pb.Name, pb.Uuid, fmt.Sprintf("启动代理：%s", pb.Name))

	util.ResponseOk(c)
}

// Stop 停止代理
func Stop(c *gin.Context, in *StopRequest) {
	// 判断代理是否存在
	pb := models.ProxyBasic{Name: in.Name}
	if err := pb.First(); err != nil {
		logger.Error("[db] proxy basic first error.", zap.Error(err))
		util.ResponseError(c, util.ErrProxyNotExist)
		return
	}

	// 停止代理
	task := serve.ProxyTask{
		ProxyBasic: pb,
	}
	if err := task.Stop(); err != nil {
		logger.Error("[sys] proxy task stop error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionStop, pb.Name, pb.Uuid, fmt.Sprintf("停止代理：%s", pb.Name))

	util.ResponseOk(c)
}

// Restart 重启代理
func Restart(c *gin.Context, in *RestartRequest) {
	// 判断代理是否存在
	pb := models.ProxyBasic{Name: in.Name}
	if err := pb.First(); err != nil {
		logger.Error("[db] proxy basic first error.", zap.Error(err))
		util.ResponseError(c, util.ErrProxyNotExist)
		return
	}

	// 启动代理
	task := serve.ProxyTask{
		ProxyBasic: pb,
	}
	if err := task.Restart(); err != nil {
		logger.Error("[sys] proxy task start error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	audit.LogWithContext(c, models.AuditModuleProxy, models.AuditActionRestart, pb.Name, pb.Uuid, fmt.Sprintf("重启代理：%s", pb.Name))

	util.ResponseOk(c)
}
