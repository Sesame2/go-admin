package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingHeader = errors.New("Authorization Header缺失")
	ErrInvalidToken  = errors.New("无效的token")
	ErrExpiredToken  = errors.New("token已过期")
	ErrParseToken    = errors.New("解析token失败")
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		// 从请求中获取Authorization
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrMissingHeader.Error(),
			})
			ctx.Abort()
			return
		}
		// 解析令牌
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization 头部格式无效",
			})
			ctx.Abort()
			return
		}
		// 验证令牌
		tokenStr := parts[1]
		claims, err := ParseToken(tokenStr, secretKey)
		if err != nil {
			errMsg := "认证失败"
			if errors.Is(err, ErrExpiredToken) {
				errMsg = "token已过期"
			} else if errors.Is(err, ErrParseToken) {
				errMsg = "无效的token"
			}
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": errMsg,
			})
			ctx.Abort()
			return
		}

		// 把用户信息存储在上下文中
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}

}

func ParseToken(tokenStr string, secretKey string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		// 检查是否过期错误
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrExpiredToken
		}
		return nil, ErrParseToken
	}

	// 验证令牌并提取声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// 生成JWT令牌
func GenerateToken(userID string, username string, role string, secretKey string, expDuration time.Duration) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// 刷新JWT令牌
func RefreshToken(tokenStr string, secretKey string, expDuration time.Duration) (string, error) {
	// 解析旧令牌而不验证过期时间
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return "", ErrParseToken
	}

	// 验证旧令牌的其他方面
	if claims, ok := token.Claims.(*Claims); ok {
		// 创建新令牌
		return GenerateToken(claims.UserID, claims.Username, claims.Role, secretKey, expDuration)
	}
	return "", ErrInvalidToken
}

// 从Gin上下文中获取userID
func GetUserIDFromContext(ctx *gin.Context) (string, error) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return "", errors.New("用户ID不存在与上下文中")
	}

	if id, ok := userID.(string); ok {
		return id, nil
	}
	return "", errors.New("用户ID类型错误")
}

// 从Gin上下文中获取用户角色
func GetRoleFromContext(ctx *gin.Context) (string, error) {
	role, exists := ctx.Get("role")
	if !exists {
		return "", errors.New("角色不存在于上下文中")
	}
	if r, ok := role.(string); ok {
		return r, nil
	}
	return "", errors.New("角色类型错误")
}

// ParseTokenWithoutValidation 解析令牌而不验证过期时间
func ParseTokenWithoutValidation(tokenStr string, secretKey string) (*Claims, error) {
	// 解析旧令牌而不验证过期时间
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, ErrParseToken
	}

	// 验证令牌并提取声明
	if claims, ok := token.Claims.(*Claims); ok {
		// 检查其他约束，但不检查过期时间
		now := time.Now()
		if claims.Issuer != "go-admin" {
			return nil, ErrInvalidToken
		}
		if claims.NotBefore != nil && now.Before(claims.NotBefore.Time) {
			return nil, ErrInvalidToken
		}

		return claims, nil
	}
	return nil, ErrInvalidToken
}
