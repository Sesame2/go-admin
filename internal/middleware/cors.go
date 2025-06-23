package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig 定义CORS配置选项
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 返回默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 返回配置好的CORS中间件
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 返回使用自定义配置的CORS中间件
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	// 规范化origins
	normalizedAllowOrigins := normalizeOrigins(config.AllowOrigins)

	// 将配置的Headers加入规范化
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := int(config.MaxAge.Seconds())

	// 返回实际的中间件函数
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查请求的Origin是否在允许列表中
		allowOrigin := getAllowOrigin(origin, normalizedAllowOrigins)

		// 设置CORS头
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
			c.Header("Access-Control-Max-Age", string(maxAge))

			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 常规请求
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Expose-Headers", exposeHeaders)

		c.Next()
	}
}

// 将Origins列表规范化
func normalizeOrigins(origins []string) []string {
	// 如果包含"*"，直接返回，表示所有源都允许
	if contains(origins, "*") {
		return []string{"*"}
	}

	// 过滤并规范化每个Origin
	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		if origin = strings.TrimSpace(origin); origin != "" {
			normalized = append(normalized, origin)
		}
	}

	return normalized
}

// 检查请求的Origin是否允许
func getAllowOrigin(origin string, allowedOrigins []string) string {
	// 所有Origin都允许
	if contains(allowedOrigins, "*") {
		return "*"
	}

	// 检查特定Origin是否在允许列表中
	if contains(allowedOrigins, origin) {
		return origin
	}

	return ""
}

// 检查列表中是否包含特定元素
func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
