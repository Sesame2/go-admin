package services

import (
	"context"

	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/google/uuid"
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

func (s *UserService) CreateUser(ctx context.Context, input *dto.CreateUserInput) (*ent.User, error){
	return s.dao.Create(ctx, input)
}
