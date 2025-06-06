// @title 用户管理系统 API
// @version 1.0
// @description 用于用户增删改查的 RESTful API
// @host localhost:8080
// @BasePath /api
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sesame2/go-admin/internal/app"
	"github.com/Sesame2/go-admin/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建应用
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}
	defer application.Close()

	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	<-quit
	log.Println("正在关闭服务器...")
}
