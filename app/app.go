package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/up-zero/my-proxy/frontend"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/middleware"
	"github.com/up-zero/my-proxy/models"
	"github.com/up-zero/my-proxy/service/info"
	"github.com/up-zero/my-proxy/service/proxy"
	"github.com/up-zero/my-proxy/service/serve"
	"github.com/up-zero/my-proxy/util"
	"go.uber.org/zap"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func router() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Cors())

	api := r.Group("/api/v1")
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": "pong",
		})
	})

	// 详细信息
	api.POST("/info", info.Info)

	auth := api.Group("/")
	auth.Use(middleware.LoginAuthCheck())

	// 代理管理
	{
		authProxy := auth.Group("/proxy")
		// 获取代理状态
		authProxy.POST("/status", BindH(proxy.Status))
		// 创建代理
		authProxy.POST("/create", BindH(proxy.Create))
		// 编辑代理
		authProxy.POST("/edit", BindH(proxy.Edit))
		// 删除代理
		authProxy.POST("/delete", BindH(proxy.Delete))
		// 启动代理
		authProxy.POST("/start", BindH(proxy.Start))
		// 停止代理
		authProxy.POST("/stop", BindH(proxy.Stop))
		// 重启代理
		authProxy.POST("/restart", BindH(proxy.Restart))
	}

	// 前端静态文件
	subFS, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		logger.Error("[sys] failed to create sub file system", zap.Error(err))
	}
	fileServer := http.FileServer(http.FS(subFS))
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			fileServer.ServeHTTP(c.Writer, c.Request)
		} else {
			// 如果是未匹配的 API 路径，返回 404
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	return r
}

// NewApp 创建服务
func NewApp(port string) {
	// 保存服务端口
	if err := (&models.ConfigBasic{}).SaveServerPort(port); err != nil {
		logger.Error("[sys] save server port error.", zap.Error(err))
		return
	}
	// 初始化代理
	serve.NewProxy()

	// 监听退出信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	// 启动服务
	go func() {
		r := router()
		server := &http.Server{
			Handler: r,
			Addr:    port,
		}
		if err := server.ListenAndServe(); err != nil {
			logger.Error(fmt.Sprintf("%s run error", util.AppName), zap.Any("ERROR", err))
			quit <- syscall.SIGINT
		}
	}()
	// 启动成功
	logger.LOGGER.Info(fmt.Sprintf("%s started successfully", util.AppName), zap.String("port", port))
	<-quit
	logger.LOGGER.Error(fmt.Sprintf("%s stopped successfully", util.AppName))
}
