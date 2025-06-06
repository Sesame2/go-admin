package dao

import (
	"context"

	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/google/uuid"
)

type UserDAO struct {
	db interfaces.Database
}

func NewUserDAO(db interfaces.Database) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) GetByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return d.db.Client().User.Get(ctx, id)
}

func (d *UserDAO) GetAll(ctx context.Context) ([]*ent.User, error) {
	return d.db.Client().User.Query().All(ctx)
}

func (d *UserDAO) Create(ctx context.Context, input *dto.CreateUserInput) (*ent.User, error) {
	return d.db.Client().User.Create().
		SetUsername(input.Username).
		SetEmail(input.Email).
		SetPassword(input.Password).
		Save(ctx)
}
