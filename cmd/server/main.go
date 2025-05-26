package main

import (
    "github.com/Sesame2/go-admin/internal/api"
)

func main() {
    
    // 设置路由
    r := api.SetupRouter()
    // api.router.SetupRoutes(r)

    // 启动服务器
    if err := r.Run(":8080"); err != nil {
        panic(err)
    }
}