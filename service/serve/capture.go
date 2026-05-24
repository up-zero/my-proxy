package serve

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

const (
	captureClientBufferSize = 256
	captureEventBufferSize  = 1024
	captureWriteWait        = 1 * time.Second
	capturePongWait         = 60 * time.Second
	capturePingPeriod       = capturePongWait * 9 / 10
	captureMaxMessageSize   = 1024
)

type captureEvent struct {
	taskUuid  string
	timestamp int64
	direction string
	protocol  string
	payload   []byte
}

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
	clients map[*captureClient]struct{}
}

type captureClient struct {
	mu         sync.RWMutex
	conn       *websocket.Conn
	send       chan []byte
	done       chan struct{}
	taskUuid   string
	remoteAddr string
	closed     bool
}

func (c *captureClient) trySend(message []byte) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false
	}

	select {
	case c.send <- message:
		return true
	default:
		return false
	}

}

func (c *captureClient) close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.send)
	close(c.done)
	c.mu.Unlock()

	_ = c.conn.Close()
}

// CaptureHub 管理抓包任务
type CaptureHub struct {
	mu       sync.RWMutex
	tasks    map[string]*clientSet // taskUuid -> set of clients
	events   chan captureEvent
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
			tasks:  make(map[string]*clientSet),
			events: make(chan captureEvent, captureEventBufferSize),
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
		go hub.run()
	})
	return hub
}

func (h *CaptureHub) run() {
	for event := range h.events {
		h.Broadcast(event.taskUuid, PacketData{
			TaskUuid:  event.taskUuid,
			Timestamp: event.timestamp,
			Direction: event.direction,
			Protocol:  event.protocol,
			Payload:   hex.EncodeToString(event.payload),
		})
	}
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("[capture] marshal packet failed", zap.Error(err), zap.String("taskUuid", taskUuid))
		return
	}

	invalidClients := make([]*captureClient, 0)

	set.RLock()
	for client := range set.clients {
		if !client.trySend(jsonData) {
			invalidClients = append(invalidClients, client)
		}
	}
	set.RUnlock()

	for _, client := range invalidClients {
		h.unregister(client)
	}
}

func (h *CaptureHub) Publish(taskUuid string, direction string, protocol string, payload []byte) {
	if len(payload) == 0 {
		return
	}

	h.mu.RLock()
	_, exists := h.tasks[taskUuid]
	h.mu.RUnlock()
	if !exists {
		return
	}

	event := captureEvent{
		taskUuid:  taskUuid,
		timestamp: time.Now().UnixMilli(),
		direction: direction,
		protocol:  protocol,
		payload:   append([]byte(nil), payload...),
	}

	select {
	case h.events <- event:
	default:
	}
}

// register 注册 WebSocket客户端
func (h *CaptureHub) register(conn *websocket.Conn, taskUuid string) *captureClient {
	client := &captureClient{
		conn:       conn,
		send:       make(chan []byte, captureClientBufferSize),
		done:       make(chan struct{}),
		taskUuid:   taskUuid,
		remoteAddr: conn.RemoteAddr().String(),
	}

	h.mu.Lock()
	set, exists := h.tasks[taskUuid]
	if !exists {
		set = &clientSet{clients: make(map[*captureClient]struct{})}
		h.tasks[taskUuid] = set
	}
	set.Lock()
	set.clients[client] = struct{}{}
	set.Unlock()
	h.mu.Unlock()

	logger.Info("[sys] capture client registered", zap.String("taskUuid", taskUuid),
		zap.String("remoteAddr", client.remoteAddr))

	return client
}

// unregister 注销 WebSocket客户端

func (h *CaptureHub) unregister(client *captureClient) {
	removed := false

	h.mu.Lock()
	set, exists := h.tasks[client.taskUuid]
	if exists {
		set.Lock()
		if _, ok := set.clients[client]; ok {
			delete(set.clients, client)
			removed = true
		}
		if len(set.clients) == 0 {
			delete(h.tasks, client.taskUuid)
		}
		set.Unlock()
	}
	h.mu.Unlock()

	client.close()

	if removed {
		logger.Info("[sys] capture client unregistered", zap.String("taskUuid", client.taskUuid),
			zap.String("remoteAddr", client.remoteAddr))
	}
}

func (h *CaptureHub) CloseTask(taskUuid string) {
	h.mu.Lock()
	set, exists := h.tasks[taskUuid]
	if !exists {
		h.mu.Unlock()
		return
	}
	delete(h.tasks, taskUuid)

	set.Lock()
	clients := make([]*captureClient, 0, len(set.clients))
	for client := range set.clients {
		clients = append(clients, client)
	}
	set.clients = make(map[*captureClient]struct{})
	set.Unlock()
	h.mu.Unlock()

	for _, client := range clients {
		client.close()
		logger.Info("[sys] capture client closed with task", zap.String("taskUuid", taskUuid),
			zap.String("remoteAddr", client.remoteAddr))
	}
}

func (h *CaptureHub) writePump(client *captureClient) {
	ticker := time.NewTicker(capturePingPeriod)
	defer func() {
		ticker.Stop()
		h.unregister(client)
	}()

	for {
		select {
		case message, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(captureWriteWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(captureWriteWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *CaptureHub) readPump(client *captureClient) {
	defer h.unregister(client)

	client.conn.SetReadLimit(captureMaxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(capturePongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(capturePongWait))
	})

	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn("[capture] websocket read failed",
					zap.Error(err),
					zap.String("taskUuid", client.taskUuid),
					zap.String("remoteAddr", client.remoteAddr))
			}
			return
		}
	}
}

func (h *CaptureHub) watchAuthExpiry(client *captureClient, expireAt int64) {
	if expireAt <= 0 {
		return
	}

	delay := time.Until(time.Unix(expireAt, 0))
	if delay <= 0 {
		h.unregister(client)
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		_ = client.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, util.MsgErrAuth),
			time.Now().Add(captureWriteWait),
		)
		h.unregister(client)
	case <-client.done:
		return
	}
}

// WebsocketCapture Websocket 连接（抓包）
func WebsocketCapture(c *gin.Context) {
	taskUuid := c.Query("task_uuid")
	if taskUuid == "" {
		util.ResponseError(c, errors.New("task uuid is required"))
		return
	}

	value, ok := proxyTaskMap.Load(taskUuid)
	if !ok {
		util.ResponseError(c, util.ErrProxyNotExist)
		return
	}

	task := value.(*ProxyTask)
	if task.State != models.ProxyStateRunning {
		util.ResponseError(c, errors.New("task is not running"))
		return
	}

	claimValue, exists := c.Get("UserClaim")
	if !exists {
		util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
		return
	}

	userClaim, ok := claimValue.(*util.UserClaim)
	if !ok || userClaim == nil || userClaim.Username == "" {
		util.ResponseMsg(c, util.CodeErrAuth, util.MsgErrAuth)
		return
	}

	// 升級HTTP
	conn, err := GetCaptureHub().upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("[capture] websocket upgrade failed", zap.Error(err))
		return
	}

	hub := GetCaptureHub()
	client := hub.register(conn, taskUuid)

	go hub.writePump(client)
	go hub.readPump(client)
	go hub.watchAuthExpiry(client, userClaim.ExpiresAt)
}
