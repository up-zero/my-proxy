package serve

import (
	"bufio"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/trafficpolicy"
	"go.uber.org/zap"
)

var proxyTaskMap = new(sync.Map)

const (
	socks5Version              = 0x05
	socks5AuthNone             = 0x00
	socks5AuthUsernamePassword = 0x02
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

	// 根据是否配置了用户名密码，选择认证方式
	requireAuth := task.Socks5Username != "" || task.Socks5Password != ""
	method := byte(socks5AuthNoAcceptable)
	if requireAuth {
		// 需要用户名密码认证
		for _, item := range methods {
			if item == socks5AuthUsernamePassword {
				method = socks5AuthUsernamePassword
				break
			}
		}
	} else {
		// 无需认证
		for _, item := range methods {
			if item == socks5AuthNone {
				method = socks5AuthNone
				break
			}
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

	// 用户名密码认证流程
	if method == socks5AuthUsernamePassword {
		if err := task.handleSocks5UsernamePasswordAuth(clientConn); err != nil {
			return err
		}
	}

	return nil
}

// handleSocks5UsernamePasswordAuth 处理 SOCKS5 用户名密码认证
func (task *ProxyTask) handleSocks5UsernamePasswordAuth(clientConn net.Conn) error {
	// version + username_len + username + password_len + password
	// version(1) + username_len(1)
	header := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, header); err != nil {
		return err
	}
	if header[0] != 0x01 {
		return fmt.Errorf("unsupported socks5 auth version(%d)", header[0])
	}

	usernameLen := int(header[1])
	if usernameLen == 0 {
		return errors.New("invalid socks5 username length (0)")
	}

	// username(N) + password_len(1)
	remainFirstPart := make([]byte, usernameLen+1)
	if _, err := io.ReadFull(clientConn, remainFirstPart); err != nil {
		return err
	}

	passwordLen := int(remainFirstPart[usernameLen])
	if passwordLen == 0 {
		return errors.New("invalid socks5 password length (0)")
	}

	// password(N)
	passwordBuf := make([]byte, passwordLen)
	if _, err := io.ReadFull(clientConn, passwordBuf); err != nil {
		return err
	}

	// username(N)
	usernameBuf := remainFirstPart[:usernameLen]

	authPayload := append(header, remainFirstPart...)
	authPayload = append(authPayload, passwordBuf...)
	task.recordPayload(&task.bytesIn, "IN", models.ProxyTypeSocks5, authPayload)

	username := string(usernameBuf)
	password := string(passwordBuf)
	authSuccess := username == task.Socks5Username && password == task.Socks5Password

	// 认证结果：version(1) + status(1)  0x00=成功
	authResp := []byte{0x01, 0x00}
	if !authSuccess {
		authResp = []byte{0x01, 0x01}
	}
	if _, err := clientConn.Write(authResp); err != nil {
		return err
	}
	task.recordPayload(&task.bytesOut, "OUT", models.ProxyTypeSocks5, authResp)

	if !authSuccess {
		return errors.New("socks5 username/password auth failed")
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

// startHttp 启动 HTTP 代理
//
// 按是否配置目标地址分为两种模式：
//   - 固定转发（填写了目标地址）：客户端直接访问本代理地址，目标由服务端指定，
//     转发时使用配置的目标地址与上游协议，Host 与上游 TLS 的 SNI 均为配置的域名，无需客户端改造
//   - 动态代理（未填写目标地址）：客户端将代理地址配置为 HTTP 代理，可动态访问任意目标地址
//
// 两种模式均支持：
//   - 普通 HTTP 请求：解析请求行中的绝对地址后转发（如 GET http://example.com/ HTTP/1.1）
//   - HTTPS 请求：通过 CONNECT 方法建立隧道转发
//   - WebSocket 等协议升级：透传 101 Switching Protocols，并桥接双向数据
func (task *ProxyTask) startHttp() error {
	transport := &http.Transport{
		// 代理直连目标，不使用系统环境变量中的代理配置
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	task.httpTransport = transport

	server := &http.Server{
		Addr:    net.JoinHostPort(task.ListenAddress, task.ListenPort),
		Handler: http.HandlerFunc(task.handleHttpProxy),
	}
	task.httpServer = server
	task.State = models.ProxyStateRunning
	proxyTaskMap.Store(task.Uuid, task)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("[sys] http proxy ListenAndServe error", zap.Error(err))
		}
	}()

	return nil
}

// handleHttpProxy HTTP 代理入口
func (task *ProxyTask) handleHttpProxy(w http.ResponseWriter, r *http.Request) {
	// 认证校验（可选）；固定转发时目标由服务端指定，客户端不会携带 Proxy-Authorization，故不校验
	if !task.IsHttpFixedForward() && !task.checkHttpProxyAuth(r) {
		w.Header().Set("Proxy-Authenticate", `Basic realm="my-proxy"`)
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return
	}

	isCapturing := task.capture.IsCapturing(task.Uuid)
	task.httpActive.Add(1)
	defer task.httpActive.Add(-1)

	if r.Method == http.MethodConnect {
		task.handleHttpConnect(w, r, isCapturing)
		return
	}
	task.handleHttpForward(w, r, isCapturing)
}

// checkHttpProxyAuth 校验 HTTP 代理认证信息（Basic 认证）
func (task *ProxyTask) checkHttpProxyAuth(r *http.Request) bool {
	if task.HttpUsername == "" && task.HttpPassword == "" {
		return true
	}
	const prefix = "Basic "
	auth := r.Header.Get("Proxy-Authorization")
	if len(auth) <= len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(auth[len(prefix):]))
	if err != nil {
		return false
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(username), []byte(task.HttpUsername)) == 1 &&
		subtle.ConstantTimeCompare([]byte(password), []byte(task.HttpPassword)) == 1
}

// handleHttpConnect 处理 CONNECT 请求（HTTPS 隧道）
func (task *ProxyTask) handleHttpConnect(w http.ResponseWriter, r *http.Request, isCapturing bool) {
	// 目标地址：固定转发使用服务端配置的目标地址；动态代理使用请求中的地址，未显式指定端口时默认 443
	targetAddr := ""
	if task.IsHttpFixedForward() {
		targetAddr = net.JoinHostPort(task.TargetAddress, task.TargetPort)
	} else {
		targetAddr = r.Host
		if targetAddr == "" {
			targetAddr = r.URL.Host
		}
		if _, _, err := net.SplitHostPort(targetAddr); err != nil {
			targetAddr = net.JoinHostPort(targetAddr, "443")
		}
	}

	if isCapturing {
		if dump, err := httputil.DumpRequest(r, false); err == nil {
			task.capture.Publish(task.Uuid, "IN", models.ProxyTypeHttp, dump)
		}
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		logger.Error("[sys] http connect target error", zap.String("target", targetAddr), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task.registerTcpConn(clientConn)
	defer task.unregisterTcpConn(clientConn)
	if buf != nil {
		clientConn = &bufferedConn{Conn: clientConn, r: buf.Reader}
	}

	resp := []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	if _, err := clientConn.Write(resp); err != nil {
		return
	}
	task.recordPayload(&task.bytesOut, "OUT", models.ProxyTypeHttp, resp)

	// 隧道数据双向转发
	done := make(chan struct{}, 2)
	go task.copyData(targetConn, clientConn, &task.bytesIn, "IN", models.ProxyTypeHttp, done)
	go task.copyData(clientConn, targetConn, &task.bytesOut, "OUT", models.ProxyTypeHttp, done)
	<-done
}

// handleHttpForward 转发普通 HTTP 请求
//
//   - 固定转发（已配置目标地址）：目标固定为服务端配置值，Host 使用配置的域名，客户端只需访问本代理
//   - 动态代理（未配置目标地址）：要求请求行为绝对地址，按请求行中的地址转发
func (task *ProxyTask) handleHttpForward(w http.ResponseWriter, r *http.Request, isCapturing bool) {
	if !task.IsHttpFixedForward() && (r.URL == nil || r.URL.Host == "") {
		// 动态代理要求请求行为绝对地址
		http.Error(w, "this is an HTTP dynamic proxy, the request url must be absolute", http.StatusBadRequest)
		return
	}

	if isCapturing {
		if dump, err := httputil.DumpRequest(r, false); err == nil {
			task.capture.Publish(task.Uuid, "IN", models.ProxyTypeHttp, dump)
		}
	}

	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	outReq.Close = false
	if task.IsHttpFixedForward() {
		// 固定转发：目标地址由服务端配置，Host（及上游 TLS 的 SNI）使用配置的域名
		scheme := task.UpstreamSchemeOrDefault()
		outReq.URL.Scheme = scheme
		outReq.URL.Host = net.JoinHostPort(task.TargetAddress, task.TargetPort)
		outReq.Host = upstreamHostHeader(task.TargetAddress, task.TargetPort, scheme)
		outReq.Header.Set("X-Forwarded-Host", r.Host)
		outReq.Header.Set("X-Forwarded-Proto", "http")
		appendXForwardedFor(outReq.Header, r.RemoteAddr)
	}
	// 协议升级（WebSocket 等）需要保留 Upgrade 相关头部
	upgradeProtocol := r.Header.Get("Upgrade")
	isUpgrade := upgradeProtocol != "" && headerValuesContainsToken(r.Header.Values("Connection"), "upgrade")
	removeHopByHopHeaders(outReq.Header)
	if isUpgrade {
		outReq.Header.Set("Connection", "Upgrade")
		outReq.Header.Set("Upgrade", upgradeProtocol)
	}

	// 包装 r.Body 并拦截数据流
	if r.Body != nil {
		outReq.Body = &httpReadCloser{
			rc:          r.Body,
			task:        task,
			isCapturing: isCapturing,
		}
	}

	resp, err := task.httpTransport.RoundTrip(outReq)
	if err != nil {
		logger.Error("[sys] http proxy round trip error", zap.String("target", outReq.URL.String()), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 协议升级：将客户端连接与目标连接桥接
	if resp.StatusCode == http.StatusSwitchingProtocols {
		task.handleHttpUpgrade(w, resp, isCapturing)
		return
	}

	removeHopByHopHeaders(resp.Header)
	copyHeader(w.Header(), resp.Header)
	task.publishHttpResponseHeader(resp.StatusCode, w.Header(), isCapturing)
	w.WriteHeader(resp.StatusCode)

	writer := &httpCountingWriter{
		w:    w,
		task: task,
	}
	// 流式响应（Content-Length 未知）实时写出，避免缓冲导致客户端等待
	if resp.ContentLength < 0 {
		if flusher, ok := w.(http.Flusher); ok {
			writer.flusher = flusher
		}
	}
	if _, err := io.Copy(writer, resp.Body); err != nil {
		logger.Error("[sys] http proxy copy response body error", zap.String("target", outReq.URL.String()), zap.Error(err))
	}
}

// handleHttpUpgrade 处理协议升级（WebSocket 等），桥接客户端与目标连接
func (task *ProxyTask) handleHttpUpgrade(w http.ResponseWriter, resp *http.Response, isCapturing bool) {
	upgradedConn, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		http.Error(w, "target response does not support protocol upgrade", http.StatusBadGateway)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	task.registerTcpConn(clientConn)
	defer task.unregisterTcpConn(clientConn)
	if buf != nil {
		clientConn = &bufferedConn{Conn: clientConn, r: buf.Reader}
	}

	// 回写 101 响应
	var respBuf bytes.Buffer
	fmt.Fprintf(&respBuf, "HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	resp.Header.Write(&respBuf)
	respBuf.WriteString("\r\n")
	if _, err := clientConn.Write(respBuf.Bytes()); err != nil {
		return
	}
	if isCapturing {
		task.capture.Publish(task.Uuid, "OUT", models.ProxyTypeHttp, respBuf.Bytes())
	}

	// 双向桥接
	done := make(chan struct{}, 2)
	go task.copyData(upgradedConn, clientConn, &task.bytesIn, "IN", models.ProxyTypeHttp, done)
	go task.copyData(clientConn, upgradedConn, &task.bytesOut, "OUT", models.ProxyTypeHttp, done)
	<-done
}

// publishHttpResponseHeader 抓包响应头（出站）
func (task *ProxyTask) publishHttpResponseHeader(statusCode int, header http.Header, isCapturing bool) {
	if !isCapturing {
		return
	}
	var headerBuf bytes.Buffer
	// 状态行
	fmt.Fprintf(&headerBuf, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))
	header.Write(&headerBuf)
	headerBuf.WriteString("\r\n") // 头部结束

	task.capture.Publish(task.Uuid, "OUT", models.ProxyTypeHttp, headerBuf.Bytes())
}

// upstreamHostHeader 生成上游请求的 Host 头（协议默认端口时省略端口）
func upstreamHostHeader(host, port, scheme string) string {
	if port == "" ||
		(scheme == models.ProxyUpstreamSchemeHttps && port == "443") ||
		(scheme == models.ProxyUpstreamSchemeHttp && port == "80") {
		return host
	}
	return net.JoinHostPort(host, port)
}

// appendXForwardedFor 将客户端地址追加到 X-Forwarded-For
func appendXForwardedFor(header http.Header, remoteAddr string) {
	clientIP := remoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		clientIP = host
	}
	if prior := header.Get("X-Forwarded-For"); prior != "" {
		header.Set("X-Forwarded-For", prior+", "+clientIP)
		return
	}
	header.Set("X-Forwarded-For", clientIP)
}

// headerValuesContainsToken 判断头部值中是否包含指定 token（大小写不敏感）
func headerValuesContainsToken(values []string, token string) bool {
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(item), token) {
				return true
			}
		}
	}
	return false
}

// removeHopByHopHeaders 移除逐跳头部（RFC 7230 6.1）
func removeHopByHopHeaders(header http.Header) {
	for _, value := range header.Values("Connection") {
		for _, name := range strings.Split(value, ",") {
			if name = strings.TrimSpace(name); name != "" {
				header.Del(name)
			}
		}
	}
	for _, name := range []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		header.Del(name)
	}
}

// copyHeader 复制头部
func copyHeader(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// bufferedConn 包装连接，优先消费 Hijack 缓冲区中已缓存的数据，避免隧道/升级时丢包
type bufferedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

// httpCountingWriter 响应体写出包装：统计出站流量并抓包
type httpCountingWriter struct {
	w       io.Writer
	task    *ProxyTask
	flusher http.Flusher
}

func (w *httpCountingWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if n > 0 {
		w.task.recordPayload(&w.task.bytesOut, "OUT", models.ProxyTypeHttp, p[:n])
		if w.flusher != nil {
			w.flusher.Flush()
		}
	}
	return n, err
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

func (task *ProxyTask) recordPayload(counter *atomic.Int64, direction string, protocol string, payload []byte) {
	if len(payload) == 0 {
		return
	}
	counter.Add(int64(len(payload)))
	if direction == "OUT" {
		task.applyTrafficPolicy(direction, int64(len(payload)))
	}
	if task.capture.IsCapturing(task.Uuid) {
		task.capture.Publish(task.Uuid, direction, protocol, payload)
	}
}

func (task *ProxyTask) applyTrafficPolicy(direction string, bytes int64) {
	if delay := trafficpolicy.RecordTraffic(trafficpolicy.TrafficEvent{
		ProxyUuid:         task.Uuid,
		Direction:         direction,
		Bytes:             bytes,
		ActiveConnections: task.ActiveConnections(),
	}); delay > 0 {
		time.Sleep(delay)
	}
}

// copyData 拷贝数据
//
// # Params:
//
//	dst: 目标写入
//	src: 源读取
//	counter: 计数器
//	direction: 数据方向，IN-入站 OUT-出站
//	protocol: 协议类型
//	done: 完成信号
func (task *ProxyTask) copyData(dst io.Writer, src io.Reader, counter *atomic.Int64, direction string, protocol string, done chan struct{}) {
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
		task.applyTrafficPolicy("OUT", int64(n))

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
	// 发送停止信号
	close(task.stopChan)
	// 关闭空闲连接
	if task.httpTransport != nil {
		task.httpTransport.CloseIdleConnections()
	}
	var err error
	if task.httpServer != nil {
		// 优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = task.httpServer.Shutdown(ctx); err != nil {
			logger.Error("[sys] http proxy shutdown error", zap.Error(err))
		}
	}
	// 关闭已建立的隧道连接（CONNECT、协议升级）
	task.mu.Lock()
	for conn := range task.tcpActiveConn {
		conn.Close()
	}
	task.mu.Unlock()
	return err
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
		// 按创建时间降序
		sort.Slice(list, func(i, j int) bool {
			return list[i].CreatedAt > list[j].CreatedAt
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
