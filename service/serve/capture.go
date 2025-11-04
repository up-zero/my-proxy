package serve

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// PacketData 推送的数据包
type PacketData struct {
	TaskUuid  string `json:"task_uuid"` // 代理唯一标识
	Timestamp int64  `json:"timestamp"` // 时间戳，毫秒
	Direction string `json:"direction"` // IN-入站 OUT-出站
	Protocol  string `json:"protocol"`  // TCP\UDP\HTTP
	Payload   string `json:"payload"`   // Hex-encoded data
}

// clientSet 客户端集合
type clientSet struct {
	sync.RWMutex
	clients map[*websocket.Conn]bool
}

// CaptureHub 管理抓包任务
type CaptureHub struct {
	mu       sync.RWMutex
	tasks    map[string]*clientSet // taskUuid -> set of clients
	upgrader websocket.Upgrader
}

var (
	hub     *CaptureHub
	hubOnce sync.Once
)

// GetCaptureHub 获取全局唯一的抓包实例
func GetCaptureHub() *CaptureHub {
	hubOnce.Do(func() {
		hub = &CaptureHub{
			tasks: make(map[string]*clientSet),
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
	})
	return hub
}

// IsCapturing 检查是否有任务在抓包
func (h *CaptureHub) IsCapturing(taskUuid string) bool {
	h.mu.RLock()
	set, exists := h.tasks[taskUuid]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	set.RLock()
	isCapturing := len(set.clients) > 0
	set.RUnlock()

	return isCapturing
}

// Broadcast 将数据推送给客户端
func (h *CaptureHub) Broadcast(taskUuid string, data PacketData) {
	h.mu.RLock()
	set, exists := h.tasks[taskUuid]
	h.mu.RUnlock()

	if !exists {
		return
	}

	jsonData, _ := json.Marshal(data)

	set.Lock()
	defer set.Unlock()

	for client := range set.clients {
		// 设置超时
		client.SetWriteDeadline(time.Now().Add(1 * time.Second))
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			// 发送失败，断开连接
			go h.unregister(client, taskUuid)
		}
	}
}

// register 注册 WebSocket客户端
func (h *CaptureHub) register(conn *websocket.Conn, taskUuid string) {
	h.mu.Lock()
	set, exists := h.tasks[taskUuid]
	if !exists {
		set = &clientSet{clients: make(map[*websocket.Conn]bool)}
		h.tasks[taskUuid] = set
	}
	h.mu.Unlock()

	set.Lock()
	set.clients[conn] = true
	set.Unlock()

	logger.Info("[sys] capture client registered", zap.String("taskUuid", taskUuid),
		zap.String("remoteAddr", conn.RemoteAddr().String()))
}

// unregister 注销 WebSocket客户端
func (h *CaptureHub) unregister(conn *websocket.Conn, taskUuid string) {
	h.mu.RLock()
	set, exists := h.tasks[taskUuid]
	h.mu.RUnlock()

	if !exists {
		return
	}

	set.Lock()
	if _, ok := set.clients[conn]; ok {
		delete(set.clients, conn)
		conn.Close()
		logger.Info("[sys] capture client unregistered", zap.String("taskUuid", taskUuid),
			zap.String("remoteAddr", conn.RemoteAddr().String()))
	}
	set.Unlock()
}

// WebsocketCapture Websocket 连接（抓包）
func WebsocketCapture(c *gin.Context) {
	taskUuid := c.Query("task_uuid")
	if taskUuid == "" {
		util.ResponseError(c, errors.New("task uuid is required"))
		return
	}

	// 升級HTTP
	conn, err := GetCaptureHub().upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("[capture] websocket upgrade failed", zap.Error(err))
		return
	}

	hub := GetCaptureHub()
	hub.register(conn, taskUuid)

	go func() {
		defer hub.unregister(conn, taskUuid)
		for {
			// 读取错误时断开连接
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
