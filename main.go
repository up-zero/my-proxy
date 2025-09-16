/*
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package main

import (
	"github.com/up-zero/my-proxy/cmd"
	"github.com/up-zero/my-proxy/logger"
	"github.com/up-zero/my-proxy/models"
)

func main() {
	// 初始化数据库
	models.NewGormDB()
	// 初始化日志
	logger.NewLogger()
	// 初始化命令行工具
	cmd.Execute()
}
