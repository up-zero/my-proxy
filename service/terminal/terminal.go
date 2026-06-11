package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

const (
	clientBufferSize = 256
	writeWait        = 1 * time.Second
	pongWait         = 60 * time.Second
	pingPeriod       = pongWait * 9 / 10
	maxMessageSize   = 4096
)

// Conn 单个终端连接
type Conn struct {
	mu         sync.RWMutex
	wsMu       sync.Mutex
	wsConn     *websocket.Conn
	sshClient  *ssh.Client
	sshSession *ssh.Session
	stdinPipe  io.WriteCloser
	send       chan []byte
	done       chan struct{} // 广播关闭信号
	language   string
	remoteAddr string
	closed     bool
}

// Hub 管理所有终端连接
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Conn // sessionId -> connection
	upgrader websocket.Upgrader
}

var (
	hub     *Hub
	hubOnce sync.Once
)

// GetHub 获取全局终端 Hub
func GetHub() *Hub {
	hubOnce.Do(func() {
		hub = &Hub{
			sessions: make(map[string]*Conn),
			upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
		}
	})
	return hub
}

// newSSHClient 创建 SSH 客户端
func newSSHClient(host, port, username, password string) (*ssh.Client, error) {
	if port == "" {
		port = "22"
	}
	addr := net.JoinHostPort(host, port)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %w", err)
	}
	return client, nil
}

// register 注册终端连接
func (h *Hub) register(wsConn *websocket.Conn, language, sessionID string) *Conn {
	tc := &Conn{
		wsConn:     wsConn,
		send:       make(chan []byte, clientBufferSize),
		done:       make(chan struct{}),
		language:   util.NormalizeLanguage(language),
		remoteAddr: wsConn.RemoteAddr().String(),
	}

	h.mu.Lock()
	h.sessions[sessionID] = tc
	h.mu.Unlock()

	logger.Info("[terminal] client registered",
		zap.String("sessionId", sessionID),
		zap.String("remoteAddr", tc.remoteAddr))
	return tc
}

// unregister 注销终端连接
func (h *Hub) unregister(tc *Conn, sessionID string) {
	h.mu.Lock()
	delete(h.sessions, sessionID)
	h.mu.Unlock()

	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return
	}
	tc.closed = true
	close(tc.done)
	tc.mu.Unlock()

	// 关闭各组件资源
	if tc.stdinPipe != nil {
		_ = tc.stdinPipe.Close()
	}
	if tc.sshSession != nil {
		_ = tc.sshSession.Close()
	}
	if tc.sshClient != nil {
		_ = tc.sshClient.Close()
	}

	// 保护并发安全地关闭 WebSocket
	tc.wsMu.Lock()
	if tc.wsConn != nil {
		_ = tc.wsConn.Close()
	}
	tc.wsMu.Unlock()

	logger.Info("[terminal] client unregistered",
		zap.String("sessionId", sessionID),
		zap.String("remoteAddr", tc.remoteAddr))
}

// safeWriteMessage 封装安全的 WebSocket 消息写入
func (tc *Conn) safeWriteMessage(messageType int, data []byte) error {
	tc.wsMu.Lock()
	defer tc.wsMu.Unlock()
	_ = tc.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return tc.wsConn.WriteMessage(messageType, data)
}

// safeWriteControl 封装安全的 WebSocket 控制帧写入
func (tc *Conn) safeWriteControl(controlType int, data []byte) error {
	tc.wsMu.Lock()
	defer tc.wsMu.Unlock()
	return tc.wsConn.WriteControl(controlType, data, time.Now().Add(writeWait))
}

// connectSSH 连接到目标 SSH 并启动 PTY
func (tc *Conn) connectSSH(host, port, username, password string, cols, rows int) error {
	client, err := newSSHClient(host, port, username, password)
	if err != nil {
		return err
	}
	tc.sshClient = client

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("create ssh session failed: %w", err)
	}
	tc.sshSession = session

	// 获取 stdin 管道
	tc.stdinPipe, err = session.StdinPipe()
	if err != nil {
		return fmt.Errorf("get stdin pipe failed: %w", err)
	}

	// 设置 stdout/stderr
	session.Stdout = &writer{tc: tc}
	session.Stderr = &writer{tc: tc}

	// 请求 PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		return fmt.Errorf("request pty failed: %w", err)
	}

	// 启动 shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("start shell failed: %w", err)
	}

	return nil
}

// resizePTY 调整终端尺寸
func (tc *Conn) resizePTY(cols, rows int) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if tc.sshSession != nil && !tc.closed {
		_ = tc.sshSession.WindowChange(rows, cols)
	}
}

// writer 将 SSH 输出写入 WebSocket
type writer struct {
	tc *Conn
}

func (w *writer) Write(p []byte) (int, error) {
	w.tc.mu.RLock()
	if w.tc.closed {
		w.tc.mu.RUnlock()
		return 0, errors.New("terminal closed")
	}
	sendChan := w.tc.send
	w.tc.mu.RUnlock()

	msg := Message{
		Type: "data",
		Data: string(p),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}

	select {
	case sendChan <- data:
		return len(p), nil
	case <-w.tc.done:
		return 0, errors.New("terminal closed")
	}
}

// writePump WebSocket 写协程
func (h *Hub) writePump(tc *Conn, sessionID string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.unregister(tc, sessionID)
	}()

	for {
		select {
		case <-tc.done: // 及时响应退出信号
			return
		case message, ok := <-tc.send:
			if !ok {
				return
			}
			if err := tc.safeWriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := tc.safeWriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump WebSocket 读协程
func (h *Hub) readPump(tc *Conn, sessionID string) {
	defer h.unregister(tc, sessionID)

	tc.wsConn.SetReadLimit(maxMessageSize)
	_ = tc.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	tc.wsConn.SetPongHandler(func(string) error {
		return tc.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := tc.wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Warn("[terminal] websocket read failed",
					zap.Error(err),
					zap.String("sessionId", sessionID),
					zap.String("remoteAddr", tc.remoteAddr))
			}
			return
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			tc.mu.RLock()
			if tc.stdinPipe != nil && !tc.closed {
				_, _ = tc.stdinPipe.Write(raw)
			}
			tc.mu.RUnlock()
			continue
		}

		switch msg.Type {
		case "resize":
			if msg.Size != nil {
				tc.resizePTY(msg.Size.Cols, msg.Size.Rows)
			}
		case "data":
			tc.mu.RLock()
			if tc.stdinPipe != nil && !tc.closed {
				_, _ = tc.stdinPipe.Write([]byte(msg.Data))
			}
			tc.mu.RUnlock()
		default:
			tc.mu.RLock()
			if tc.stdinPipe != nil && !tc.closed {
				_, _ = tc.stdinPipe.Write(raw)
			}
			tc.mu.RUnlock()
		}
	}
}

// watchAuthExpiry 监听 token 过期
func (h *Hub) watchAuthExpiry(tc *Conn, sessionID string, expireAt int64) {
	if expireAt <= 0 {
		return
	}

	delay := time.Until(time.Unix(expireAt, 0))
	if delay <= 0 {
		h.unregister(tc, sessionID)
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		// 多线程安全控制写入
		_ = tc.safeWriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, util.TranslateMessage(tc.language, util.MsgErrAuth)),
		)
		h.unregister(tc, sessionID)
	case <-tc.done:
		return
	}
}

// WebsocketTerminal WebSocket 连接（终端）
func WebsocketTerminal(c *gin.Context) {
	host := c.Query("host")
	port := c.Query("port")
	username := c.Query("username")
	password := c.Query("password")
	sessionID := c.Query("session_id")

	if host == "" || username == "" || sessionID == "" {
		util.ResponseError(c, errors.New("missing parameters"))
		return
	}
	if port == "" {
		port = "22"
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

	conn, err := GetHub().upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("[terminal] websocket upgrade failed", zap.Error(err))
		return
	}

	h := GetHub()
	tc := h.register(conn, util.GetRequestLanguage(c), sessionID)

	cols, rows := 80, 24

	if err := tc.connectSSH(host, port, username, password, cols, rows); err != nil {
		logger.Error("[terminal] ssh connect failed", zap.Error(err), zap.String("host", host), zap.String("sessionId", sessionID))
		errMsg := Message{
			Type: "error",
			Data: fmt.Sprintf("SSH connection failed: %v", err),
		}
		if data, e := json.Marshal(errMsg); e == nil {
			_ = conn.WriteMessage(websocket.TextMessage, data)
		}
		h.unregister(tc, sessionID)
		return
	}

	// 启动监控协程，监听远程 Linux 服务的正常/异常退出状态
	go func() {
		if tc.sshSession != nil {
			_ = tc.sshSession.Wait() // 阻塞直到远程终端断开
		}
		h.unregister(tc, sessionID)
	}()

	go h.writePump(tc, sessionID)
	go h.readPump(tc, sessionID)
	go h.watchAuthExpiry(tc, sessionID, userClaim.ExpiresAt)
}
