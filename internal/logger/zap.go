package logger

import (
	"os"
	"sync"

	"github.com/Sesame2/go-admin/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger全局日志实例
	Logger *zap.Logger
	once   sync.Once
)

func Setup(cfg *config.Config) {
	once.Do(
		func() {
			var err error
			Logger, err = buildLogger(cfg)
			if err != nil {
				os.Stderr.WriteString("无法初始化日志：" + err.Error() + "\n")
				os.Exit(1)
			}
		},
	)
}

func buildLogger(cfg *config.Config) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Server.Mode == "release" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logLevel := cfg.Logger.Level
	level := zap.InfoLevel // 默认info级别
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// 创建日志实例
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}
	return logger, nil
}
