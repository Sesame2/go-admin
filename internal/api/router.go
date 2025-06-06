package api

import (
	"net/http"

	_ "github.com/Sesame2/go-admin/docs"
	"github.com/Sesame2/go-admin/internal/api/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CallRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "欢迎访问用户管理系统 API",
		"version": "1.0.0",
	})
}

func SetupRouter(userController controller.UserController) *gin.Engine {
	r := gin.Default()

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", CallRoot)
	// // 用户相关路由
	userGroup := r.Group("/api/users")
	{
		userGroup.GET("/", userController.GetAllUser)
		userGroup.GET("/:id", userController.GetUser)
		userGroup.POST("/", userController.CreateUser)
	}
	return r
}
