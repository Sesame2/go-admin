package services

import (
	"time"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/errors"
	"github.com/Sesame2/go-admin/internal/middleware"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var tempRole = "admin"

type AuthService struct {
	userDAO *dao.UserDAO
	config  *config.Config
}

func NewAuthService(userDAO *dao.UserDAO, config *config.Config) *AuthService {
	return &AuthService{
		userDAO: userDAO,
		config:  config,
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
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, tempRole, s.config.JWT.SecretKey, time.Duration(s.config.JWT.ExpirationHours)*time.Hour)
	if err != nil {
		return "", err
	}
	return token, nil
}
