package dao

import (
	"context"

	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/Sesame2/go-admin/internal/models/ent/user"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserDAO struct {
	db     interfaces.Database
	logger *zap.Logger
}

func NewUserDAO(db interfaces.Database, logger *zap.Logger) *UserDAO {
	daoLogger := logger.With(zap.String("component", "UserDAO"))
	return &UserDAO{
		db:     db,
		logger: daoLogger,
	}
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
		SetRole(input.Role).
		Save(ctx)
}

func (d *UserDAO) UpdateUser(ctx context.Context, id uuid.UUID, input *dto.UpdateUserInput) (*ent.User, error) {
	update := d.db.Client().User.UpdateOneID(id)
	// 有条件地应用更新
	if input.Username != nil {
		update.SetUsername(*input.Username)
	}
	if input.Email != nil {
		update.SetEmail(*input.Email)
	}
	if input.Password != nil {
		update.SetPassword(*input.Password)
	}

	return update.Save(ctx)
}

func (d *UserDAO) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	return d.db.Client().User.Query().Where(user.ID(id)).Exist(ctx)
}

func (d *UserDAO) GetByUserName(ctx context.Context, username string) (*ent.User, error) {
	return d.db.Client().User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
}
