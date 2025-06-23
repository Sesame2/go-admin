package services

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/errors"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	dao    *dao.UserDAO
	logger *zap.Logger
}

func NewUserService(dao *dao.UserDAO, logger *zap.Logger) *UserService {
	logger = logger.With(zap.String("component", "Userservice"))
	return &UserService{
		dao:    dao,
		logger: logger,
	}
}

func (s *UserService) ListUsers(ctx context.Context) ([]*ent.User, error) {
	s.logger.Info("获取所有用户")
	return s.dao.GetAll(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	exist, err := s.dao.Exist(ctx, id)
	if err != nil {
		s.logger.Error("检查用户是否存在出错", zap.Error(err))
		return nil, fmt.Errorf("检查用户是否存在出错：%w", err)
	}
	if !exist {
		return nil, errors.ErrUserNotFound
	}
	return s.dao.GetByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, input *dto.CreateUserInput) (*ent.User, error) {
	// 增加密码加密
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败：%w", err)
		}
		passwordStr := string(hashedPassword)
		input.Password = passwordStr
	}
	user, err := s.dao.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败：%w", err)
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input *dto.UpdateUserInput) (*ent.User, error) {
	exist, err := s.dao.Exist(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("检查用户是否存在出错：%w", err)
	}
	if !exist {
		return nil, errors.ErrUserNotFound
	}
	if input.Password != nil {
		hasdedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败：%w", err)
		}
		passwordStr := string(hasdedPassword)
		input.Password = &passwordStr
	}
	return s.dao.UpdateUser(ctx, id, input)
}
