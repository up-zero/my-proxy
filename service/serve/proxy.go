package serve

import (
	"context"
	"errors"
	"fmt"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"sync"
	"time"
)

var proxyTaskMap = new(sync.Map)

// Start 启动任务
func (task *ProxyTask) Start() error {
	// 初始化 task
	task.stopChan = make(chan struct{})

	// 启动服务
	var err error
	switch task.Type {
	case models.ProxyTypeTcp:
		err = task.startTcp()
	case models.ProxyTypeUdp:
		err = task.startUdp()
	case models.ProxyTypeHttp:
		err = task.startHttp()
	default:
		err = fmt.Errorf("proxy type(%s) not support", task.Type)
	}
	return err
}

// startTcp 启动TCP代理
func (task *ProxyTask) startTcp() error {
	task.tcpActiveConn = make(map[net.Conn]struct{})
	listener, err := net.Listen("tcp", ":"+task.ListenPort)
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
	go task.copyData(targetConn, clientConn, done)
	// target -> client
	go task.copyData(clientConn, targetConn, done)
	<-done
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

	listenAddr := ":" + task.ListenPort
	server := &http.Server{
		Addr:    listenAddr,
		Handler: proxy, // 将处理器设置为我们的反向代理
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

func (task *ProxyTask) copyData(dst, src net.Conn, done chan struct{}) {
	defer func() {
		select {
		case done <- struct{}{}:
		default:
		}
	}()
	io.Copy(dst, src)
}

func (task *ProxyTask) registerTcpConn(conn net.Conn) {
	task.mu.Lock()
	defer task.mu.Unlock()
	task.tcpActiveConn[conn] = struct{}{}
}

func (task *ProxyTask) unregisterTcpConn(conn net.Conn) {
	conn.Close()
	task.mu.Lock()
	defer task.mu.Unlock()
	delete(task.tcpActiveConn, conn)
}

func (task *ProxyTask) startUdp() error {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+task.ListenPort)
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
				logger.Error("[sys] read from udp error.", zap.Error(err))
				continue
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
	default:
		err = fmt.Errorf("proxy type(%s) not support", task.Type)
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
		return []*ProxyTask{value.(*ProxyTask)}, nil
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
		return list, nil
	}
}

// Remove 移除任务
func (task *ProxyTask) Remove() {
	proxyTaskMap.Delete(task.Uuid)
}
