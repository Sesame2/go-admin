package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery

		// 处理请求
		ctx.Next()

		cost := time.Since(start)

		// 记录请求日志
		logger.Debug(
			"HTTP请求",
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("ip", ctx.ClientIP()),
			zap.Duration("latency", cost),
			zap.Int("size", ctx.Writer.Size()),
			zap.String("user-agent", ctx.Request.UserAgent()),
			zap.String("errors", ctx.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
