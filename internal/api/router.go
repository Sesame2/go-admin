package api

import (
	"net/http"

	_ "github.com/Sesame2/go-admin/docs"
	"github.com/Sesame2/go-admin/internal/api/controller"
	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/middleware"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// API服务结构
type API struct {
	UserController *controller.UserController
	AuthController *controller.AuthController
	Config         *config.Config
}

func NewAPI(config *config.Config, userController *controller.UserController, authController *controller.AuthController) *API {
	return &API{
		UserController: userController,
		AuthController: authController,
		Config:         config,
	}
}

func CallRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "欢迎访问用户管理系统 API",
		"version": "1.0.0",
	})
}

func (api *API) SetupRouter(logger *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.LoggerMiddleware(logger))
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", CallRoot)
	// // 用户相关路由
	userGroup := r.Group("/api/users")
	authGroup := r.Group("/api/auth")
	userGroup.Use(middleware.JWTAuthMiddleware(api.Config.JWT.SecretKey))
	{
		userGroup.GET("/", api.UserController.GetAllUser)
		userGroup.GET("/:id", api.UserController.GetUser)
		userGroup.POST("/", api.UserController.CreateUser)
		userGroup.PUT("/:id", api.UserController.UpdateUser)

		authGroup.POST("/login", api.AuthController.Login)
	}
	
	
	return r
}
