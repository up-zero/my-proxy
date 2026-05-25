package proxy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/gotool/sliceutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

func loadProxyGroupMap() (map[string]string, error) {
	list := make([]*models.GroupBasic, 0)
	if err := models.DB.Model(new(models.GroupBasic)).Find(&list).Error; err != nil {
		logger.Error("[db] get group list error.", zap.Error(err))
		return nil, err
	}
	groupMap := make(map[string]string, len(list))
	for _, item := range list {
		groupMap[item.Uuid] = item.Name
	}
	return groupMap, nil
}

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

func normalizeProxyBasic(pb *models.ProxyBasic) {
	pb.Name = strings.TrimSpace(pb.Name)
	pb.GroupUuid = strings.TrimSpace(pb.GroupUuid)
	pb.Type = strings.ToUpper(strings.TrimSpace(pb.Type))
	pb.ListenAddress = strings.TrimSpace(pb.ListenAddress)
	pb.ListenPort = strings.TrimSpace(pb.ListenPort)
	pb.TargetAddress = strings.TrimSpace(pb.TargetAddress)
	pb.TargetPort = strings.TrimSpace(pb.TargetPort)

	if pb.Type == models.ProxyTypeSocks5 {
		pb.TargetAddress = ""
		pb.TargetPort = ""
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
	groupMap, err := loadProxyGroupMap()
	if err != nil {
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	for _, item := range tasks {
		if pb, ok := proxyMap[item.Uuid]; ok {
			item.GroupUuid = pb.GroupUuid
			item.GroupName = groupMap[pb.GroupUuid]
		}
	}
	if in.Name != "" {
		tasks = sliceutil.Filter(tasks, func(pt *serve.ProxyTask) bool {
			return strings.Contains(pt.Name, in.Name)
		})
	}
	if in.GroupUuid != "" {
		tasks = sliceutil.Filter(tasks, func(pt *serve.ProxyTask) bool {
			return pt.GroupUuid == in.GroupUuid
		})
	}
	util.ResponseOkWithList(c, tasks)
}

// savePreValid 判断代理信息是否有效
func savePreValid(pb *models.ProxyBasic) error {
	normalizeProxyBasic(pb)
	if err := validateProxyBasicFields(pb); err != nil {
		return err
	}
	if pb.GroupUuid != "" {
		if err := (&models.GroupBasic{Uuid: pb.GroupUuid}).First(); err != nil {
			logger.Error("[db] get group basic error.", zap.Error(err))
			return fmt.Errorf("group does not exist")
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
	if err := models.DB.Create(pb).Error; err != nil {
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
	if err := models.DB.Model(new(models.ProxyBasic)).Where("uuid = ?", pb.Uuid).Updates(map[string]any{
		"name":           pb.Name,
		"group_uuid":     pb.GroupUuid,
		"type":           pb.Type,
		"listen_address": pb.ListenAddress,
		"listen_port":    pb.ListenPort,
		"target_address": pb.TargetAddress,
		"target_port":    pb.TargetPort,
		"state":          pb.State,
	}).Error; err != nil {
		logger.Error("[sys] proxy basic updates error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

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
		if err := create(pb); err != nil {
			logger.Error("[sys] proxy basic create error.", zap.Error(err))
			util.ResponseError(c, err)
			return
		}
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
	if err := models.DB.Delete(new(models.ProxyBasic), "uuid = ?", pb.Uuid).Error; err != nil {
		logger.Error("[sys] proxy basic delete error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

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
	if err := models.DB.Delete(new(models.ProxyBasic), "uuid in ?", in.Uuid).Error; err != nil {
		logger.Error("[sys] proxy basic delete error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
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

	util.ResponseOk(c)
}
