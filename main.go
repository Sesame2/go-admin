package main

import (
    "github.com/gin-gonic/gin"
    "log"
)

func main() {
    r := gin.Default()

    // 这里可以添加路由，后续可以从 api 目录引入路由配置
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ROOT CALLED!",
		})
	})

    // 启动服务器
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("failed to start server: %v", err)
    }
}