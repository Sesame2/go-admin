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
	DocumentController      *controller.DocumentController
	Config                  *config.Config
}

func NewAPI(
	config *config.Config,
	userController *controller.UserController,
	authController *controller.AuthController,
	kbController *controller.KnowledgeBaseController,
	docController *controller.DocumentController,
) *API {
	return &API{
		UserController:          userController,
		AuthController:          authController,
		KnowledgeBaseController: kbController,
		DocumentController:      docController,
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

	// API 路由组
	apiGroup := r.Group("/api")

	// 公开路由（不需要认证）
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/login", api.AuthController.Login)
		authGroup.POST("/refresh", api.AuthController.Refresh)
	}

	// 需要认证的路由
	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(middleware.JWTAuthMiddleware(api.Config.JWT.SecretKey))
	{
		// 用户管理
		userGroup := protectedGroup.Group("/users")
		{
			userGroup.GET("", api.UserController.GetAllUser)
			userGroup.GET("/:id", api.UserController.GetUser)
			userGroup.POST("", api.UserController.CreateUser)
			userGroup.PUT("/:id", api.UserController.UpdateUser)
			userGroup.DELETE("/:id", api.UserController.DeleteUser)
		}

		// 知识库管理
		kbGroup := protectedGroup.Group("/knowledge_bases")
		{
			kbGroup.POST("", api.KnowledgeBaseController.CreateKnowledgeBase)
			kbGroup.GET("", api.KnowledgeBaseController.GetAllKnowledgeBase)
			kbGroup.GET("/:id", api.KnowledgeBaseController.GetKnowledgeBaseByID)
			kbGroup.PUT("/:id", api.KnowledgeBaseController.UpdateKnowledgeBase)
			kbGroup.DELETE("/:id", api.KnowledgeBaseController.DeleteKnowledgeBase)
		}

		// 文档管理
		docGroup := protectedGroup.Group("/documents")
		{
			docGroup.POST("/upload", api.DocumentController.UploadDocument)
			docGroup.POST("/parse", api.DocumentController.ParseDocument)
			docGroup.GET("", api.DocumentController.ListDocuments)
			docGroup.GET("/:id", api.DocumentController.GetDocument)
			docGroup.GET("/:id/download", api.DocumentController.GetDocumentDownloadURL)
			docGroup.PUT("/:id", api.DocumentController.UpdateDocument)
			docGroup.DELETE("/:id", api.DocumentController.DeleteDocument)
		}

		// 检索问答
		retrievalGroup := protectedGroup.Group("/retrieval")
		{
			retrievalGroup.GET("/stream", api.DocumentController.StreamRetrieve)
		}
	}

	return r
}
