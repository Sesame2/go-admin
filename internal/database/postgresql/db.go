package postgresql

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type PostgreSQL struct {
	db *gorm.DB
}

func NewDatabase(cfg config.DatabaseConfig) (interfaces.Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败：%w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败：%w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return &PostgreSQL{db: db}, nil
}

// DB 返回 GORM 数据库实例
func (p *PostgreSQL) DB() *gorm.DB {
	return p.db
}

// Close 关闭数据库连接
func (p *PostgreSQL) Close() error {
	if p.db == nil {
		return nil
	}
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate 执行数据库迁移
func (p *PostgreSQL) Migrate(ctx context.Context) error {
	// 启用 pgvector 扩展
	p.db.Exec("CREATE EXTENSION IF NOT EXISTS vector")

	// 创建 document_status ENUM 类型（与 Python 模型保持一致）
	p.db.Exec(`
		DO $$ BEGIN
			CREATE TYPE document_status AS ENUM (
				'pending',
				'downloading', 
				'parsing',
				'chunking',
				'tagging',
				'embedding',
				'completed',
				'failed'
			);
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`)

	// 自动迁移所有模型
	err := p.db.WithContext(ctx).AutoMigrate(
		&models.User{},
		&models.KnowledgeBase{},
		&models.Document{},
		&models.DocumentChunk{},
		&models.AtomQuestion{},
	)
	if err != nil {
		return fmt.Errorf("数据库迁移失败：%w", err)
	}
	return nil
}
