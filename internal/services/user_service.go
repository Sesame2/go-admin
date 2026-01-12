package services

import (
	"context"
	"errors"
	"fmt"

	customerrors "github.com/Sesame2/go-admin/internal/errors"
	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
	log  *zap.Logger
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
		log:  logger.NewModuleLogger("UserService"),
	}
}

func (s *UserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	s.log.Info("获取所有用户")
	return s.repo.GetAll(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	exist, err := s.repo.Exist(ctx, id)
	if err != nil {
		s.log.Error("检查用户是否存在出错", zap.Error(err))
		return nil, fmt.Errorf("检查用户是否存在出错：%w", err)
	}
	if !exist {
		return nil, customerrors.ErrUserNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, input *dto.CreateUserInput) (*models.User, error) {
	// 增加密码加密
	hashedPassword := input.Password
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败：%w", err)
		}
		hashedPassword = string(hash)
	}

	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
		Role:     input.Role,
	}

	err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败：%w", err)
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, input *dto.UpdateUserInput) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, customerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("获取用户失败：%w", err)
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败：%w", err)
		}
		user.Password = string(hashedPassword)
	}

	err = s.repo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("更新用户失败：%w", err)
	}
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	exist, err := s.repo.Exist(ctx, id)
	if err != nil {
		return fmt.Errorf("检查用户是否存在出错：%w", err)
	}
	if !exist {
		return customerrors.ErrUserNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, customerrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
