package interfaces

import (
	"context"

	"gorm.io/gorm"
)

// Database 数据库接口
type Database interface {
	DB() *gorm.DB
	Close() error
	Migrate(ctx context.Context) error
}
