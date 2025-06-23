package app

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/api"
	"github.com/Sesame2/go-admin/internal/api/controller"
	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/database"
	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Application struct {
	DB             interfaces.Database
	Router         *gin.Engine
	Config         *config.Config
	UserController *controller.UserController
	Logger         *zap.Logger
}

// New 创建一个新的应用实例
func New(cfg *config.Config) (*Application, error) {
	// 初始化日志系统
	logger.Setup(cfg)
	log := logger.Logger

	log.Info("初始化应用程序...")
	// 初始化数据库
	db, err := database.Factory(cfg.Database)
	if err != nil {
		return nil, err
	}

	// 执行数据库迁移
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		return nil, err
	}
	log.Info("数据库迁移完成")

	// 初始化DAO层
	userDAO := dao.NewUserDAO(db, log)

	// 初始化服务层
	userService := services.NewUserService(userDAO, log)
	authService := services.NewAuthService(userDAO, cfg, log)

	// 初始化控制器层
	userController := controller.NewUserController(userService, log)
	authController := controller.NewAuthController(authService, log)

	// 初始化api服务
	apiService := api.NewAPI(cfg, userController, authController)

	// 初始化路由
	router := apiService.SetupRouter(log)

	// 设置生产环境下的模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	return &Application{
		DB:             db,
		Router:         router,
		Config:         cfg,
		UserController: userController,
	}, nil
}

// Run 启动应用服务
func (app *Application) Run() error {
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	return app.Router.Run(addr)
}

// Close 关闭应用，释放资源
func (app *Application) Close() error {
	if app.DB != nil {
		return app.DB.Close()
	}
	return nil
}
