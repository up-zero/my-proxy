package proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// Status 获取代理状态
func Status(c *gin.Context, in *StatusRequest) {
	task := serve.ProxyTask{}
	if in.Name != "" && in.Name != "all" {
		// 获取单条任务
		pb := models.ProxyBasic{Name: in.Name}
		if err := pb.First(); err != nil {
			logger.Error("[db] proxy basic first error.", zap.Error(err))
			util.ResponseError(c, util.ErrProxyNotExist)
			return
		}
		task.Uuid = pb.Uuid
	}
	tasks, err := task.Status()
	if err != nil {
		logger.Error("[sys] proxy task status error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	util.ResponseOkWithList(c, tasks)
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

	// 名称判重
	count, err := pb.CountForName()
	if err != nil {
		logger.Error("[sys] proxy basic count for name error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	if count > 0 {
		util.ResponseError(c, util.ErrNameExists)
		return
	}
	// 端口判重
	count, err = pb.CountForPort()
	if err != nil {
		logger.Error("[sys] proxy basic count for port error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	if count > 0 {
		util.ResponseError(c, util.ErrListenPortExists)
		return
	}

	// 启动代理任务
	task := serve.ProxyTask{
		ProxyBasic: *pb,
	}
	if err := task.Start(); err != nil {
		logger.Error("[sys] proxy task start error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}

	// 落库
	if err := models.DB.Create(pb).Error; err != nil {
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

	// 名称判重
	count, err := pb.CountForName()
	if err != nil {
		logger.Error("[sys] proxy basic count for name error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	if count > 0 {
		util.ResponseError(c, util.ErrNameExists)
		return
	}
	// 端口判重
	count, err = pb.CountForPort()
	if err != nil {
		logger.Error("[sys] proxy basic count for port error.", zap.Error(err))
		util.ResponseError(c, err)
		return
	}
	if count > 0 {
		util.ResponseError(c, util.ErrListenPortExists)
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
	if err := models.DB.Model(pb).Where("uuid = ?", pb.Uuid).Updates(pb).Error; err != nil {
		logger.Error("[sys] proxy basic updates error.", zap.Error(err))
		util.ResponseError(c, err)
		return
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
