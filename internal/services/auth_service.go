package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/dao"
	customerrors "github.com/Sesame2/go-admin/internal/errors"
	"github.com/Sesame2/go-admin/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userDAO *dao.UserDAO
	config  *config.Config
	logger  *zap.Logger
}

func NewAuthService(userDAO *dao.UserDAO, config *config.Config, logger *zap.Logger) *AuthService {
	logger = logger.With(zap.String("component", "AuthService"))
	return &AuthService{
		userDAO: userDAO,
		config:  config,
		logger:  logger,
	}
}

func (s *AuthService) Login(ctx *gin.Context, username, password string) (string, error) {
	user, err := s.userDAO.GetByUserName(ctx, username)
	if err != nil {
		return "", customerrors.ErrUserNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", customerrors.ErrInvalidCredentials
	}
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, user.Role, s.config.JWT.SecretKey, time.Duration(s.config.JWT.ExpirationHours)*time.Hour)
	if err != nil {
		return "", err
	}
	s.logger.Info(fmt.Sprintln("用户", username, "登录"))
	return token, nil
}

func (s *AuthService) Refresh(ctx *gin.Context, tokenStr string) (string, error) {
	// 验证当前令牌
	claims, err := middleware.ParseToken(tokenStr, s.config.JWT.SecretKey)
	if err != nil {
		if errors.Is(err, middleware.ErrExpiredToken) {
			// 令牌过期但仍在允许的刷新时间内，可以继续刷新
			// 解析令牌而不验证有效期
			claims, err = middleware.ParseTokenWithoutValidation(tokenStr, s.config.JWT.SecretKey)
			if err != nil {
				s.logger.Error("解析令牌失败", zap.Error(err))
				return "", customerrors.ErrParseToken
			}
		} else {
			s.logger.Error("无效的令牌，尝试刷新", zap.Error(err))
			return "", customerrors.ErrInvalidToken
		}
	}

	// 2. 检查用户是否存在
	userID := claims.UserID
	userUUID, err := uuid.Parse(userID) // 确保userID是有效的UUID格式
	if err != nil {
		s.logger.Error("无效的用户ID格式", zap.String("userID", userID), zap.Error(err))
		return "", customerrors.ErrInvalidUserIDFormat
	}
	exists, err := s.userDAO.Exist(ctx, userUUID)
	if err != nil {
		s.logger.Error("检查用户是否存在失败", zap.String("userID", userID), zap.Error(err))
		return "", err
	}
	if !exists {
		s.logger.Error("用户不存在", zap.String("userID", userID))
		return "", customerrors.ErrUserNotFound
	}
	user, err := s.userDAO.GetByID(ctx, userUUID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.String("userID", userID), zap.Error(err))
		return "", err
	}
	// 3. 生成新的令牌
	newToken, err := middleware.GenerateToken(user.ID.String(), user.Username, user.Role, s.config.JWT.SecretKey, time.Duration(s.config.JWT.ExpirationHours)*time.Hour)
	if err != nil {
		s.logger.Error("生成新的令牌失败", zap.Error(err))
		return "", err
	}
	s.logger.Info(fmt.Sprintf("用户 %s 刷新令牌成功", user.Username), zap.String("userID", user.ID.String()))
	return newToken, nil
}
