package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
)

// 不应被代理转发的路径前缀（始终由主节点本地处理）
var localOnlyPathPrefixes = []string{
	"/api/v1/login",
	"/api/v1/refresh/token",
	"/api/v1/edit/password",
	"/api/v1/node/",
	"/api/v1/ws/",
}

// 转发到子节点时不应携带的请求头（逐跳头 + 可能引起冲突的头）
var skipForwardHeaders = map[string]bool{
	"authorization":       true,
	"x-node-id":           true,
	"host":                true,
	"origin":              true,
	"referer":             true,
	"content-length":      true,
	"transfer-encoding":   true,
	"connection":          true,
	"accept-encoding":     true,
	"te":                  true,
	"trailer":             true,
	"upgrade":             true,
	"proxy-authorization": true,
	"proxy-authenticate":  true,
}

// NodeProxy 节点代理中间件：检测 X-Node-Id 请求头，若非本地节点则将请求转发到目标节点
func NodeProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		nodeId := strings.TrimSpace(c.GetHeader("X-Node-Id"))
		if nodeId == "" || nodeId == "node-local" {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// 排除本地专用路由
		for _, prefix := range localOnlyPathPrefixes {
			if strings.HasPrefix(path, prefix) || path == prefix {
				c.Next()
				return
			}
		}

		// 查询节点信息
		node := &models.NodeBasic{Uuid: nodeId}
		if err := node.First(); err != nil {
			logger.Warn("[node-proxy] node not found", zap.String("node_id", nodeId), zap.Error(err))
			c.Next()
			return
		}

		if !node.Enabled {
			logger.Warn("[node-proxy] node is disabled", zap.String("node_id", nodeId))
			c.Next()
			return
		}

		// 生成超管 token（使用目标节点的密钥）
		claim := &util.UserClaim{
			Username: "admin",
			Level:    models.UserLevelRoot,
			RoleID:   "role-admin",
		}
		token, err := claim.GenerateTokenWithKey(time.Now().Add(30*time.Second).Unix(), node.SecretKey)
		if err != nil {
			logger.Error("[node-proxy] generate token error", zap.String("node_id", nodeId), zap.Error(err))
			c.Next()
			return
		}

		// 读取原始请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 构建目标 URL
		targetURL := "http://" + node.Address + path
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		// 构建转发请求（仅携带安全头）
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			logger.Error("[node-proxy] create proxy request error", zap.String("url", targetURL), zap.Error(err))
			c.Next()
			return
		}

		// 选择性复制请求头：跳过逐跳头及可能冲突的头
		for key, values := range c.Request.Header {
			if skipForwardHeaders[strings.ToLower(key)] {
				continue
			}
			for _, v := range values {
				proxyReq.Header.Add(key, v)
			}
		}

		// 设置代理所需的头
		proxyReq.Header.Set("Authorization", token)
		proxyReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))
		proxyReq.Header.Set("Accept", "application/json")

		// 发送请求（禁用自动重定向以透传状态码）
		client := &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Do(proxyReq)
		if err != nil {
			logger.Error("[node-proxy] proxy request error", zap.String("url", targetURL), zap.Error(err))
			util.ResponseMsg(c, util.CodeErr, util.MsgErrNet)
			c.Abort()
			return
		}
		defer resp.Body.Close()

		// 读取响应体
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("[node-proxy] read response body error", zap.Error(err))
			util.ResponseMsg(c, util.CodeErr, util.MsgErrNet)
			c.Abort()
			return
		}

		// 仅回写安全的响应头（Content-Type 等，跳过 CORS / 逐跳头）
		for key, values := range resp.Header {
			keyLower := strings.ToLower(key)
			if strings.HasPrefix(keyLower, "access-control-") ||
				keyLower == "content-length" ||
				keyLower == "content-encoding" ||
				keyLower == "transfer-encoding" ||
				keyLower == "connection" {
				continue
			}
			for _, v := range values {
				c.Header(key, v)
			}
		}

		// 回写响应
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
		c.Abort()
	}
}
