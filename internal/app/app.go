package app

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/api"
	"github.com/Sesame2/go-admin/internal/api/controller"
	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/database"
	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/mq/rabbitmq"
	"github.com/Sesame2/go-admin/internal/repository"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/Sesame2/go-admin/internal/storage/s3"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Application struct {
	DB       interfaces.Database
	Router   *gin.Engine
	Config   *config.Config
	MQClient *rabbitmq.RabbitMQClient
	log      *zap.Logger
}

// New 创建一个新的应用实例
func New(cfg *config.Config) (*Application, error) {
	// 初始化日志系统
	logger.Setup(cfg)
	log := logger.NewModuleLogger("App")

	log.Info("初始化应用程序...")

	// 初始化数据库
	db, err := database.Factory(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 执行数据库迁移
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	log.Info("数据库迁移完成")

	// 初始化MinIO客户端
	var minioClient *s3.MinioClient
	if cfg.S3.Endpoint != "" {
		minioClient, err = s3.NewMinioClient(cfg)
		if err != nil {
			log.Warn("初始化MinIO客户端失败，文件上传功能将不可用", zap.Error(err))
		} else {
			log.Info("MinIO客户端初始化成功")
		}
	}

	// 初始化RabbitMQ客户端
	var mqClient *rabbitmq.RabbitMQClient
	if cfg.RabbitMQ.Host != "" {
		mqClient = rabbitmq.NewRabbitMQClient(&cfg.RabbitMQ, logger.Logger)
		if err := mqClient.Connect(); err != nil {
			log.Warn("连接RabbitMQ失败，消息队列功能将不可用", zap.Error(err))
			mqClient = nil
		} else {
			log.Info("RabbitMQ连接成功")
		}
	}

	// 初始化Repository层
	userRepo := repository.NewUserRepository(db)
	kbRepo := repository.NewKnowledgeBaseRepository(db)
	docRepo := repository.NewDocumentRepository(db)

	// 初始化服务层
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, cfg)
	kbService := services.NewKnowledgeBaseService(kbRepo)
	docService := services.NewDocumentService(docRepo, kbRepo, minioClient, mqClient)

	// 初始化检索服务
	var retrievalService *services.RetrievalService
	if cfg.RAGService.BaseURL != "" {
		retrievalService = services.NewRetrievalService(cfg.RAGService.BaseURL)
		log.Info("检索服务初始化成功", zap.String("base_url", cfg.RAGService.BaseURL))
	}

	// 初始化控制器层
	userController := controller.NewUserController(userService)
	authController := controller.NewAuthController(authService)
	kbController := controller.NewKnowledgeBaseController(kbService)
	docController := controller.NewDocumentController(docService, retrievalService)

	// 初始化api服务
	apiService := api.NewAPI(cfg, userController, authController, kbController, docController)

	// 初始化路由
	router := apiService.SetupRouter()

	// 设置生产环境下的模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Info("应用程序初始化完成")

	return &Application{
		DB:       db,
		Router:   router,
		Config:   cfg,
		MQClient: mqClient,
		log:      log,
	}, nil
}

// Run 启动应用服务
func (app *Application) Run() error {
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	app.log.Info("启动HTTP服务", zap.String("addr", addr))
	return app.Router.Run(addr)
}

// Close 关闭应用，释放资源
func (app *Application) Close() error {
	app.log.Info("关闭应用程序...")

	if app.MQClient != nil {
		if err := app.MQClient.Close(); err != nil {
			app.log.Error("关闭RabbitMQ连接失败", zap.Error(err))
		}
	}

	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			app.log.Error("关闭数据库连接失败", zap.Error(err))
			return err
		}
	}

	app.log.Info("应用程序已关闭")
	return nil
}
