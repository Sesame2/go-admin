package api

import (
	"github.com/Sesame2/go-admin/internal/api/controller"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine{
	r := gin.Default()
	userController := controller.NewUserController()
	// 用户相关路由
	userGroup := r.Group("/api/users")
	{
		userGroup.GET("/", userController.GetUser)
	}
	return r
}
