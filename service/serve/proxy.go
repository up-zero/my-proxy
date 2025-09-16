package serve

import (
	"errors"
	"fmt"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"go.uber.org/zap"
	"io"
	"net"
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
	task.mu.Lock()
	task.tcpActiveConn[clientConn] = struct{}{}
	task.mu.Unlock()

	defer func() {
		clientConn.Close()
		task.mu.Lock()
		delete(task.tcpActiveConn, clientConn)
		task.mu.Unlock()
	}()

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

func (task *ProxyTask) copyData(dst, src net.Conn, done chan struct{}) {
	defer func() {
		select {
		case done <- struct{}{}:
		default:
		}
	}()
	io.Copy(dst, src)
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
