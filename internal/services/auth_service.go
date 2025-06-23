package services

import (
	"fmt"
	"time"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/errors"
	"github.com/Sesame2/go-admin/internal/middleware"
	"github.com/gin-gonic/gin"
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
		return "", errors.ErrUserNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.ErrInvalidCredentials
	}
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, user.Role, s.config.JWT.SecretKey, time.Duration(s.config.JWT.ExpirationHours)*time.Hour)
	if err != nil {
		return "", err
	}
	s.logger.Info(fmt.Sprintln("用户", username, "登录"))
	return token, nil
}
