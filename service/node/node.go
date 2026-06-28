package node

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/up-zero/gotool/idutil"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// List 获取节点列表
func List(c *gin.Context) {
	nodes, err := (&models.NodeBasic{}).All()
	if err != nil {
		logger.Error("[node] list error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	items := make([]NodeItem, 0, len(nodes))
	for _, n := range nodes {
		items = append(items, NodeItem{
			Uuid:      n.Uuid,
			Name:      n.Name,
			Address:   n.Address,
			SecretKey: n.SecretKey,
			Enabled:   n.Enabled,
			IsLocal:   n.IsLocal,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		})
	}
	util.ResponseOkWithData(c, &ListResponse{List: items})
}

// Create 创建节点
func Create(c *gin.Context, in *CreateRequest) {
	// 名称判重
	cnt, err := (&models.NodeBasic{Name: in.Name}).CountForName()
	if err != nil {
		logger.Error("[node] count for name error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	if cnt > 0 {
		util.ResponseMsg(c, util.CodeErr, util.LocalizeMessage(c, util.MsgErrNameExist))
		return
	}

	node := &models.NodeBasic{
		Uuid:      idutil.UUIDGenerate(),
		Name:      strings.TrimSpace(in.Name),
		Address:   strings.TrimSpace(in.Address),
		SecretKey: strings.TrimSpace(in.SecretKey),
		Enabled:   true,
		IsLocal:   false,
	}
	if in.Enabled != nil {
		node.Enabled = *in.Enabled
	}
	if err := node.Create(); err != nil {
		logger.Error("[node] create error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Update 更新节点
func Update(c *gin.Context, in *UpdateRequest) {
	node := &models.NodeBasic{Uuid: in.Uuid}
	if err := node.First(); err != nil {
		logger.Error("[node] find error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDataNotExist, util.MsgErrDataNotExist)
		return
	}

	// Local 节点不允许修改名称
	if !node.IsLocal {
		cnt, err := (&models.NodeBasic{Uuid: in.Uuid, Name: in.Name}).CountForName()
		if err != nil {
			logger.Error("[node] count for name error.", zap.Error(err))
			util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
			return
		}
		if cnt > 0 {
			util.ResponseMsg(c, util.CodeErr, util.LocalizeMessage(c, util.MsgErrNameExist))
			return
		}
		node.Name = strings.TrimSpace(in.Name)
	}

	node.Address = strings.TrimSpace(in.Address)
	node.SecretKey = strings.TrimSpace(in.SecretKey)
	if in.Enabled != nil {
		node.Enabled = *in.Enabled
	}

	if err := node.Update(); err != nil {
		logger.Error("[node] update error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}

// Delete 删除节点
func Delete(c *gin.Context, in *DeleteRequest) {
	node := &models.NodeBasic{Uuid: in.Uuid}
	if err := node.First(); err != nil {
		logger.Error("[node] find error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDataNotExist, util.MsgErrDataNotExist)
		return
	}
	// Local 节点不允许删除
	if node.IsLocal {
		util.ResponseMsg(c, util.CodeErr, util.LocalizeMessage(c, "cannot delete local node"))
		return
	}
	if err := node.Delete(); err != nil {
		logger.Error("[node] delete error.", zap.Error(err))
		util.ResponseMsg(c, util.CodeErrDB, util.MsgErrDB)
		return
	}
	util.ResponseOk(c)
}
