package api

import (
	"net/http"

	_ "github.com/Sesame2/go-admin/docs"
	"github.com/Sesame2/go-admin/internal/api/controller"
	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// API服务结构
type API struct {
	UserController          *controller.UserController
	AuthController          *controller.AuthController
	KnowledgeBaseController *controller.KnowledgeBaseController
	Config                  *config.Config
}

func NewAPI(config *config.Config, userController *controller.UserController, authController *controller.AuthController, kbController *controller.KnowledgeBaseController) *API {
	return &API{
		UserController:          userController,
		AuthController:          authController,
		KnowledgeBaseController: kbController,
		Config:                  config,
	}
}

func CallRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "欢迎访问用户管理系统 API",
		"version": "1.0.0",
	})
}

func (api *API) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.LoggerMiddleware())
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", CallRoot)
	// // 用户相关路由
	userGroup := r.Group("/api/users")
	authGroup := r.Group("/api/auth")
	knowledgebaseGroup := r.Group("/api/knowledge_bases")
	userGroup.Use(middleware.JWTAuthMiddleware(api.Config.JWT.SecretKey))
	{
		userGroup.GET("/", api.UserController.GetAllUser)
		userGroup.GET("/:id", api.UserController.GetUser)
		userGroup.POST("/", api.UserController.CreateUser)
		userGroup.PUT("/:id", api.UserController.UpdateUser)
		userGroup.DELETE("/:id", api.UserController.DeleteUser)

		authGroup.POST("/login", api.AuthController.Login)
		authGroup.POST("/refresh", api.AuthController.Refresh)

		knowledgebaseGroup.POST("/", api.KnowledgeBaseController.CreateKnowledgeBase)
		knowledgebaseGroup.GET("/", api.KnowledgeBaseController.GetAllKnowledgeBase)
		knowledgebaseGroup.GET("/:id", api.KnowledgeBaseController.GetKnowledgeBaseByID)
		knowledgebaseGroup.PUT("/:id", api.KnowledgeBaseController.UpdateKnowledgeBase)
		knowledgebaseGroup.DELETE("/:id", api.KnowledgeBaseController.DeleteKnowledgeBase)
	}

	return r
}
