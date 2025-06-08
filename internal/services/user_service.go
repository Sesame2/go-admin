package services

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	dao *dao.UserDAO
}

func NewUserService(dao *dao.UserDAO) *UserService {
	return &UserService{dao: dao}
}

func (s *UserService) ListUsers(ctx context.Context) ([]*ent.User, error) {
	return s.dao.GetAll(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*ent.User, error) {
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
