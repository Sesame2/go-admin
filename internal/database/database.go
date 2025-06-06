package database

import (
	"fmt"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/database/postgresql"
)

func Factory(cfg config.DatabaseConfig) (interfaces.Database, error) {
	switch cfg.Driver {
	case "postgres":
		return postgresql.NewDatabase(cfg)
	default:
		return nil, fmt.Errorf("不支持的数据类型: %s", cfg.Driver)
	}
}

