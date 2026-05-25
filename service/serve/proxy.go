package serve

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"go.uber.org/zap"
)

var proxyTaskMap = new(sync.Map)

const (
	socks5Version              = 0x05
	socks5AuthNone             = 0x00
	socks5AuthNoAcceptable     = 0xFF
	socks5CmdConnect           = 0x01
	socks5AtypIPv4             = 0x01
	socks5AtypDomain           = 0x03
	socks5AtypIPv6             = 0x04
	socks5ReplySucceeded       = 0x00
	socks5ReplyGeneralFailure  = 0x01
	socks5ReplyCommandNotAllow = 0x07
	socks5ReplyAddrNotAllow    = 0x08
)

func isNetClosedError(err error) bool {
	return errors.Is(err, net.ErrClosed)
}

// Start 启动任务
func (task *ProxyTask) Start() error {
	// 初始化 task
	task.stopChan = make(chan struct{})
	task.capture = GetCaptureHub()
	task.bytesIn.Store(0)
	task.bytesOut.Store(0)
	task.httpActive.Store(0)

	// 启动服务
	var err error
	switch task.Type {
	case models.ProxyTypeTcp:
		err = task.startTcp()
	case models.ProxyTypeUdp:
		err = task.startUdp()
	case models.ProxyTypeHttp:
		err = task.startHttp()
	case models.ProxyTypeSocks5:
		err = task.startSocks5()
	default:
		err = fmt.Errorf("proxy type(%s) not support", task.Type)
	}
	return err
}

// startTcp 启动TCP代理
func (task *ProxyTask) startTcp() error {
	task.tcpActiveConn = make(map[net.Conn]struct{})
	listener, err := net.Listen("tcp", net.JoinHostPort(task.ListenAddress, task.ListenPort))
	if err != nil {
		logger.Error("[sys] proxy task start error", zap.Error(err))
		return err
	}
	task.tcpListener = listener
	task.State = models.ProxyStateRunning
	proxyTaskMap.Store(task.Uuid, task)

	// 监听连接
	go func() {
		for {
			select {
			case <-task.stopChan:
				return
			default:
				if listener, ok := task.tcpListener.(*net.TCPListener); ok {
					listener.SetDeadline(time.Now().Add(1 * time.Second))
				}
				clientConn, err := listener.Accept()
				if err != nil {
					var opErr *net.OpError
					if errors.As(err, &opErr) && opErr.Timeout() {
						continue
					}
					if isNetClosedError(err) {
						return
					}
					logger.Error("[sys] proxy task accept error", zap.Error(err))
					continue
				}
				go task.handleConnection(clientConn)
			}
		}
	}()

	return nil
}

func (task *ProxyTask) handleConnection(clientConn net.Conn) {
	task.registerTcpConn(clientConn)
	defer task.unregisterTcpConn(clientConn)

	targetConn, err := net.DialTimeout("tcp", task.TargetAddress+":"+task.TargetPort, 5*time.Second)
	if err != nil {
		logger.Error("[sys] proxy task connect target error", zap.Error(err))
		return
	}
	defer targetConn.Close()

	done := make(chan struct{}, 2)
	// client -> target
	go task.copyData(targetConn, clientConn, &task.bytesIn, "IN", models.ProxyTypeTcp, done)
	// target -> client
	go task.copyData(clientConn, targetConn, &task.bytesOut, "OUT", models.ProxyTypeTcp, done)
	<-done
}

func (task *ProxyTask) startSocks5() error {
	task.tcpActiveConn = make(map[net.Conn]struct{})
	listener, err := net.Listen("tcp", net.JoinHostPort(task.ListenAddress, task.ListenPort))
	if err != nil {
		logger.Error("[sys] socks5 proxy task start error", zap.Error(err))
		return err
	}
	task.tcpListener = listener
	task.State = models.ProxyStateRunning
	proxyTaskMap.Store(task.Uuid, task)

	go func() {
		for {
			select {
			case <-task.stopChan:
				return
			default:
				if listener, ok := task.tcpListener.(*net.TCPListener); ok {
					listener.SetDeadline(time.Now().Add(1 * time.Second))
				}
				clientConn, err := listener.Accept()
				if err != nil {
					var opErr *net.OpError
					if errors.As(err, &opErr) && opErr.Timeout() {
						continue
					}
					if isNetClosedError(err) {
						return
					}
					logger.Error("[sys] socks5 proxy task accept error", zap.Error(err))
					continue
				}
				go task.handleSocks5Connection(clientConn)
			}
		}
	}()

	return nil
}

func (task *ProxyTask) handleSocks5Connection(clientConn net.Conn) {
	task.registerTcpConn(clientConn)
	defer task.unregisterTcpConn(clientConn)

	if err := task.handleSocks5Greeting(clientConn); err != nil {
		logger.Error("[sys] socks5 greeting error", zap.Error(err))
		return
	}

	targetConn, err := task.handleSocks5Connect(clientConn)
	if err != nil {
		logger.Error("[sys] socks5 connect error", zap.Error(err))
		return
	}
	defer targetConn.Close()

	done := make(chan struct{}, 2)
	go task.copyData(targetConn, clientConn, &task.bytesIn, "IN", models.ProxyTypeSocks5, done)
	go task.copyData(clientConn, targetConn, &task.bytesOut, "OUT", models.ProxyTypeSocks5, done)
	<-done
}

func (task *ProxyTask) handleSocks5Greeting(clientConn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, header); err != nil {
		return err
	}
	if header[0] != socks5Version {
		return fmt.Errorf("unsupported socks version(%d)", header[0])
	}

	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(clientConn, methods); err != nil {
		return err
	}
	task.recordPayload(&task.bytesIn, "IN", models.ProxyTypeSocks5, append(header, methods...))

	method := byte(socks5AuthNoAcceptable)
	for _, item := range methods {
		if item == socks5AuthNone {
			method = socks5AuthNone
			break
		}
	}

	resp := []byte{socks5Version, method}
	if _, err := clientConn.Write(resp); err != nil {
		return err
	}
	task.recordPayload(&task.bytesOut, "OUT", models.ProxyTypeSocks5, resp)
	if method == socks5AuthNoAcceptable {
		return errors.New("socks5 auth method not supported")
	}

	return nil
}

func (task *ProxyTask) handleSocks5Connect(clientConn net.Conn) (net.Conn, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, header); err != nil {
		return nil, err
	}
	if header[0] != socks5Version {
		return nil, fmt.Errorf("unsupported socks version(%d)", header[0])
	}
	if header[1] != socks5CmdConnect {
		if err := task.writeSocks5Reply(clientConn, socks5ReplyCommandNotAllow, nil); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("socks5 command(%d) not supported", header[1])
	}

	host, addrRaw, err := readSocks5Address(clientConn, header[3])
	if err != nil {
		if writeErr := task.writeSocks5Reply(clientConn, socks5ReplyAddrNotAllow, nil); writeErr != nil {
			return nil, writeErr
		}
		return nil, err
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBuf); err != nil {
		return nil, err
	}
	requestPayload := append(append(header, addrRaw...), portBuf...)
	task.recordPayload(&task.bytesIn, "IN", models.ProxyTypeSocks5, requestPayload)

	port := binary.BigEndian.Uint16(portBuf)
	targetConn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), 5*time.Second)
	if err != nil {
		if writeErr := task.writeSocks5Reply(clientConn, socks5ReplyGeneralFailure, nil); writeErr != nil {
			return nil, writeErr
		}
		return nil, err
	}

	if err := task.writeSocks5Reply(clientConn, socks5ReplySucceeded, targetConn.LocalAddr()); err != nil {
		targetConn.Close()
		return nil, err
	}

	return targetConn, nil
}

func readSocks5Address(r io.Reader, atyp byte) (string, []byte, error) {
	switch atyp {
	case socks5AtypIPv4:
		buf := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", nil, err
		}
		return net.IP(buf).String(), buf, nil
	case socks5AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(r, lenBuf); err != nil {
			return "", nil, err
		}
		hostBuf := make([]byte, int(lenBuf[0]))
		if _, err := io.ReadFull(r, hostBuf); err != nil {
			return "", nil, err
		}
		return string(hostBuf), append(lenBuf, hostBuf...), nil
	case socks5AtypIPv6:
		buf := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", nil, err
		}
		return net.IP(buf).String(), buf, nil
	default:
		return "", nil, fmt.Errorf("socks5 atyp(%d) not supported", atyp)
	}
}

func (task *ProxyTask) writeSocks5Reply(clientConn net.Conn, reply byte, addr net.Addr) error {
	resp := buildSocks5Reply(reply, addr)
	if _, err := clientConn.Write(resp); err != nil {
		return err
	}
	task.recordPayload(&task.bytesOut, "OUT", models.ProxyTypeSocks5, resp)
	return nil
}

func buildSocks5Reply(reply byte, addr net.Addr) []byte {
	resp := []byte{socks5Version, reply, 0x00}
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		if ip4 := tcpAddr.IP.To4(); ip4 != nil {
			resp = append(resp, socks5AtypIPv4)
			resp = append(resp, ip4...)
		} else if ip16 := tcpAddr.IP.To16(); ip16 != nil {
			resp = append(resp, socks5AtypIPv6)
			resp = append(resp, ip16...)
		} else {
			resp = append(resp, socks5AtypIPv4, 0, 0, 0, 0)
		}
		portBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(portBuf, uint16(tcpAddr.Port))
		resp = append(resp, portBuf...)
		return resp
	}

	resp = append(resp, socks5AtypIPv4, 0, 0, 0, 0, 0, 0)
	return resp
}

func (task *ProxyTask) startHttp() error {
	scheme := "http"
	if task.TargetPort == "443" {
		scheme = "https"
	}
	targetURLStr := fmt.Sprintf("%s://%s", scheme, task.TargetAddress)
	// 仅在端口不是协议默认端口时才追加端口号
	if task.TargetPort != "" {
		if (scheme == "https" && task.TargetPort != "443") || (scheme == "http" && task.TargetPort != "80") {
			targetURLStr += ":" + task.TargetPort
		}
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		logger.Error("[sys] reverse proxy invalid target URL", zap.String("url", targetURLStr), zap.Error(err))
		return err
	}

	// 创建一个反向代理处理器
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// 修改Host头
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = targetURL.Host
	}

	// HTTP 中间件
	handler := task.httpMiddleware(proxy)

	server := &http.Server{
		Addr:    net.JoinHostPort(task.ListenAddress, task.ListenPort),
		Handler: handler, // 将处理器设置为 http middleware
	}
	task.httpServer = server
	task.State = models.ProxyStateRunning
	proxyTaskMap.Store(task.Uuid, task)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("[sys] reverse http proxy ListenAndServe error", zap.Error(err))
		}
	}()

	return nil
}

func (task *ProxyTask) httpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isCapturing := task.capture.IsCapturing(task.Uuid)
		task.httpActive.Add(1)
		defer task.httpActive.Add(-1)

		// 抓包请求头（入站）
		if isCapturing {
			if dump, err := httputil.DumpRequest(r, false); err == nil {
				task.capture.Publish(task.Uuid, "IN", models.ProxyTypeHttp, dump)
			}
		}

		// 包装 r.Body 并拦截数据流
		if r.Body != nil {
			r.Body = &httpReadCloser{
				rc:          r.Body,
				task:        task,
				isCapturing: isCapturing,
			}
		}

		next.ServeHTTP(&httpResponseWriter{
			ResponseWriter: w,
			task:           task,
			isCapturing:    isCapturing,
		}, r)
	})
}

type httpReadCloser struct {
	rc          io.ReadCloser
	task        *ProxyTask
	isCapturing bool
}

func (r *httpReadCloser) Read(p []byte) (int, error) {
	n, err := r.rc.Read(p)
	if n > 0 {
		// 统计（入站）
		r.task.bytesIn.Add(int64(n))

		// 抓包（入站）
		if r.isCapturing {
			r.task.capture.Publish(r.task.Uuid, "IN", models.ProxyTypeHttp, p[:n])
		}
	}
	return n, err
}

func (r *httpReadCloser) Close() error {
	return r.rc.Close()
}

type httpResponseWriter struct {
	http.ResponseWriter
	task        *ProxyTask
	isCapturing bool
	headersSent bool
	statusCode  int
}

func (w *httpResponseWriter) WriteHeader(statusCode int) {
	if w.headersSent {
		return
	}
	w.headersSent = true
	w.statusCode = statusCode

	if w.isCapturing {
		var headerBuf bytes.Buffer
		// 状态行
		fmt.Fprintf(&headerBuf, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))
		w.Header().Write(&headerBuf)
		headerBuf.WriteString("\r\n") // 头部结束

		w.task.capture.Publish(w.task.Uuid, "OUT", models.ProxyTypeHttp, headerBuf.Bytes())
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *httpResponseWriter) Write(b []byte) (int, error) {
	if !w.headersSent {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(b)
	if n > 0 {
		// 统计（出站）
		w.task.bytesOut.Add(int64(n))

		// 抓包（出站）
		if w.isCapturing {
			w.task.capture.Publish(w.task.Uuid, "OUT", models.ProxyTypeHttp, b[:n])
		}
	}
	return n, err
}

// Flush 流式输出
func (w *httpResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack 处理器接管底层连接，用于 WebSocket、协议升级、低级连接等
func (w *httpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not implement Hijacker: %T", w.ResponseWriter)
	}
	w.headersSent = true
	return h.Hijack()
}

func (task *ProxyTask) recordPayload(counter *atomic.Int64, direction string, protocol string, payload []byte) {
	if len(payload) == 0 {
		return
	}
	counter.Add(int64(len(payload)))
	if task.capture.IsCapturing(task.Uuid) {
		task.capture.Publish(task.Uuid, direction, protocol, payload)
	}
}

// copyData 拷贝数据
//
// # Params:
//
//	dst: 目标连接
//	src: 源连接
//	counter: 计数器
//	direction: 数据方向，IN-入站 OUT-出站
//	done: 完成信号
func (task *ProxyTask) copyData(dst, src net.Conn, counter *atomic.Int64, direction string, protocol string, done chan struct{}) {
	defer func() {
		select {
		case done <- struct{}{}:
		default:
		}
	}()

	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// 写入数据
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				task.recordPayload(counter, direction, protocol, buf[0:nw])
			}
			if ew != nil {
				break
			}
			if nr != nw {
				break // 写入字节不匹配
			}
		}
		if er != nil {
			break // 读取错误
		}
	}
}

func (task *ProxyTask) registerTcpConn(conn net.Conn) {
	task.mu.Lock()
	defer task.mu.Unlock()
	if task.tcpActiveConn == nil {
		task.tcpActiveConn = make(map[net.Conn]struct{})
	}
	task.tcpActiveConn[conn] = struct{}{}
}

func (task *ProxyTask) unregisterTcpConn(conn net.Conn) {
	conn.Close()
	task.mu.Lock()
	defer task.mu.Unlock()
	delete(task.tcpActiveConn, conn)
}

func (task *ProxyTask) startUdp() error {
	udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(task.ListenAddress, task.ListenPort))
	if err != nil {
		logger.Error("[sys] udp proxy failed to resolve listen address", zap.Error(err))
		return err
	}
	listenConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Error("[sys] udp proxy task start error", zap.Error(err))
		return err
	}
	task.udpListener = listenConn
	task.udpSessions = make(map[string]net.Conn)
	task.State = models.ProxyStateRunning
	proxyTaskMap.Store(task.Uuid, task)

	// 监听客户端请求
	go func() {
		buffer := make([]byte, 65535) // Max UDP packet size
		for {
			select {
			case <-task.stopChan:
				return
			default:
			}

			// 1秒超时
			task.udpListener.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, clientAddr, err := task.udpListener.ReadFromUDP(buffer)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}
				if isNetClosedError(err) {
					return
				}
				logger.Error("[sys] read from udp error.", zap.Error(err))
				continue
			}

			// 统计（入站）
			task.bytesIn.Add(int64(n))

			// 抓包（入站）
			if task.capture.IsCapturing(task.Uuid) {
				task.capture.Publish(task.Uuid, "IN", models.ProxyTypeUdp, buffer[:n])
			}

			clientAddrStr := clientAddr.String()
			task.mu.Lock()
			targetConn, found := task.udpSessions[clientAddrStr]
			if !found {
				targetUDPAddr, err := net.ResolveUDPAddr("udp", task.TargetAddress+":"+task.TargetPort)
				if err != nil {
					logger.Error("[sys] udp proxy failed to resolve target address", zap.Error(err))
					task.mu.Unlock()
					continue
				}

				newTargetConn, err := net.DialUDP("udp", nil, targetUDPAddr)
				if err != nil {
					logger.Error("[sys] udp proxy failed to dial target", zap.Error(err))
					task.mu.Unlock()
					continue
				}

				targetConn = newTargetConn
				task.udpSessions[clientAddrStr] = targetConn

				// target -> client
				go task.handleUdpResponse(clientAddr, newTargetConn)
			}
			task.mu.Unlock()

			// client -> target
			if _, err := targetConn.Write(buffer[:n]); err != nil {
				logger.Error("[sys] udp failed to write to target", zap.Error(err))
			}
		}
	}()

	return nil
}

// handleUdpResponse 将target响应的内容转发到client
func (task *ProxyTask) handleUdpResponse(clientAddr net.Addr, targetConn *net.UDPConn) {
	const udpSessionTimeout = 60 * time.Second

	defer func() {
		targetConn.Close()
		task.mu.Lock()
		delete(task.udpSessions, clientAddr.String())
		task.mu.Unlock()
	}()

	respBuffer := make([]byte, 65535)
	for {
		targetConn.SetReadDeadline(time.Now().Add(udpSessionTimeout))
		n, err := targetConn.Read(respBuffer)
		if err != nil {
			return
		}

		// 统计（出站）
		task.bytesOut.Add(int64(n))

		// 抓包（出站）
		if task.capture.IsCapturing(task.Uuid) {
			task.capture.Publish(task.Uuid, "OUT", models.ProxyTypeUdp, respBuffer[:n])
		}

		_, err = task.udpListener.WriteTo(respBuffer[:n], clientAddr)
		if err != nil {
			return
		}
	}
}

// Stop 停止代理
func (task *ProxyTask) Stop() error {
	value, ok := proxyTaskMap.Load(task.Uuid)
	if !ok {
		return fmt.Errorf("task(%s) not found", task.Name)
	}
	proxyTask := value.(*ProxyTask)
	if proxyTask.State != models.ProxyStateRunning {
		return fmt.Errorf("task(%s) is not running", task.Name)
	}
	proxyTask.State = models.ProxyStateStopped

	var err error
	switch proxyTask.Type {
	case models.ProxyTypeTcp:
		err = proxyTask.stopTcp()
	case models.ProxyTypeUdp:
		err = proxyTask.stopUdp()
	case models.ProxyTypeHttp:
		err = proxyTask.stopHttpProxy()
	case models.ProxyTypeSocks5:
		err = proxyTask.stopTcp()
	default:
		err = fmt.Errorf("proxy type(%s) not support", task.Type)
	}

	if err == nil && proxyTask.capture != nil {
		proxyTask.capture.CloseTask(proxyTask.Uuid)
	}

	return err
}

// stopTcp 停止TCP代理
func (task *ProxyTask) stopTcp() error {
	// 发送停止信号
	close(task.stopChan)
	// 关闭监听器
	if task.tcpListener != nil {
		task.tcpListener.Close()
	}

	// 关闭存在的连接
	task.mu.Lock()
	for conn, _ := range task.tcpActiveConn {
		conn.Close()
	}
	task.mu.Unlock()

	return nil
}

func (task *ProxyTask) stopUdp() error {
	// 发送停止信号
	close(task.stopChan)
	// 关闭监听器
	if task.udpListener != nil {
		task.udpListener.Close()
	}

	// 关闭存在的连接
	task.mu.Lock()
	defer task.mu.Unlock()
	for _, conn := range task.udpSessions {
		conn.Close()
	}

	return nil
}

func (task *ProxyTask) stopHttpProxy() error {
	close(task.stopChan)
	if task.httpServer != nil {
		// 优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := task.httpServer.Shutdown(ctx); err != nil {
			logger.Error("[sys] reverse http proxy shutdown error", zap.Error(err))
			return err
		}
	}
	return nil
}

// Restart 重启任务
func (task *ProxyTask) Restart() error {
	task.Stop()
	return task.Start()
}

// Status 获取任务状态
func (task *ProxyTask) Status() ([]*ProxyTask, error) {
	if task.Uuid != "" {
		// 获取单条任务
		value, ok := proxyTaskMap.Load(task.Uuid)
		if !ok {
			return nil, fmt.Errorf("task(%s) not found", task.Name)
		}
		proxyTask := value.(*ProxyTask)
		proxyTask.TrafficIn = proxyTask.bytesIn.Load()
		proxyTask.TrafficOut = proxyTask.bytesOut.Load()
		return []*ProxyTask{proxyTask}, nil
	} else {
		// 获取所有任务
		list := make([]*ProxyTask, 0)
		proxyTaskMap.Range(func(key, value interface{}) bool {
			list = append(list, value.(*ProxyTask))
			return true
		})
		// 按名称升序
		sort.Slice(list, func(i, j int) bool {
			return list[i].Name < list[j].Name
		})
		for _, item := range list {
			item.TrafficIn = item.bytesIn.Load()
			item.TrafficOut = item.bytesOut.Load()
		}
		return list, nil
	}
}

type TaskSnapshot struct {
	models.ProxyBasic
	TrafficIn         int64 `json:"traffic_in"`
	TrafficOut        int64 `json:"traffic_out"`
	ActiveConnections int64 `json:"active_connections"`
}

func (task *ProxyTask) ActiveConnections() int64 {
	task.mu.Lock()
	tcpConnCnt := len(task.tcpActiveConn)
	udpSessionCnt := len(task.udpSessions)
	task.mu.Unlock()

	return int64(tcpConnCnt+udpSessionCnt) + task.httpActive.Load()
}

func (task *ProxyTask) Snapshot() *TaskSnapshot {
	return &TaskSnapshot{
		ProxyBasic:        task.ProxyBasic,
		TrafficIn:         task.bytesIn.Load(),
		TrafficOut:        task.bytesOut.Load(),
		ActiveConnections: task.ActiveConnections(),
	}
}

func ProxyTaskSnapshots() []*TaskSnapshot {
	list := make([]*TaskSnapshot, 0)
	proxyTaskMap.Range(func(key, value interface{}) bool {
		if task, ok := value.(*ProxyTask); ok {
			list = append(list, task.Snapshot())
		}
		return true
	})

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

// Remove 移除任务
func (task *ProxyTask) Remove() {
	GetCaptureHub().CloseTask(task.Uuid)
	proxyTaskMap.Delete(task.Uuid)
}
