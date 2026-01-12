package logger

import (
	"fmt"
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

// customLevelEncoder 自定义日志级别编码器，格式为 [LEVEL]
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", level.CapitalString()))
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t interface{ Format(string) string }, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05"))
}

func buildLogger(cfg *config.Config) (*zap.Logger, error) {
	// 创建自定义编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "", // 不显示调用者信息
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据模式选择输出格式
	var encoder zapcore.Encoder
	if cfg.Server.Mode == "release" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志级别
	logLevel := cfg.Logger.Level
	level := zap.InfoLevel
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

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 创建日志实例
	logger := zap.New(core)
	return logger, nil
}

// NewModuleLogger 创建带模块名的日志记录器
// 输出格式: [LEVEL] [ModuleName] message
func NewModuleLogger(moduleName string) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.Named(moduleName)
}
