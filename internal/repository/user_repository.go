package repository

import (
	"context"
	"errors"

	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrNotFound 记录未找到错误
var ErrNotFound = errors.New("record not found")

// UserRepository 用户仓储
type UserRepository struct {
	db  interfaces.Database
	log *zap.Logger
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db interfaces.Database) *UserRepository {
	return &UserRepository{
		db:  db,
		log: logger.NewModuleLogger("UserRepository"),
	}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.DB().WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Debug("用户不存在", zap.String("id", id.String()))
			return nil, ErrNotFound
		}
		r.log.Error("获取用户失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

// GetAll 获取所有用户
func (r *UserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	err := r.db.DB().WithContext(ctx).Find(&users).Error
	if err != nil {
		r.log.Error("获取所有用户失败", zap.Error(err))
		return nil, err
	}
	r.log.Info("获取所有用户成功", zap.Int("count", len(users)))
	return users, nil
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	err := r.db.DB().WithContext(ctx).Create(user).Error
	if err != nil {
		r.log.Error("创建用户失败", zap.String("username", user.Username), zap.Error(err))
		return err
	}

	r.log.Info("创建用户成功", zap.String("id", user.ID.String()), zap.String("username", user.Username))
	return nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	err := r.db.DB().WithContext(ctx).Save(user).Error
	if err != nil {
		r.log.Error("更新用户失败", zap.String("id", user.ID.String()), zap.Error(err))
		return err
	}

	r.log.Info("用户更新成功", zap.String("id", user.ID.String()))
	return nil
}

// Exist 检查用户是否存在
func (r *UserRepository) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.DB().WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		r.log.Error("检查用户存在失败", zap.String("id", id.String()), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.DB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Debug("用户不存在", zap.String("username", username))
			return nil, ErrNotFound
		}
		r.log.Error("根据用户名获取用户失败", zap.String("username", username), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB().WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
	if err != nil {
		r.log.Error("删除用户失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	r.log.Info("删除用户成功", zap.String("id", id.String()))
	return nil
}
