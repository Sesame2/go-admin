package postgresql

import (
	"context"
	"fmt"
	"log"

	"github.com/Sesame2/go-admin/internal/config"
	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/models/ent"
	_ "github.com/lib/pq"
)

type PostgreSQL struct {
	client *ent.Client
}

func NewDatabase(cfg config.DatabaseConfig) (interfaces.Database, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败：%v", err)
		return nil, fmt.Errorf("连接数据失败：%w", err)
	}

	// FIXME 不知道为什么无法设置连接池 因为无法获取 client.DB()方法

	log.Println("成功连接到PostgreSQL数据库")
	return &PostgreSQL{client: client}, nil
}

// 实现Client方法，返回ent客户端
func (p *PostgreSQL) Client() *ent.Client {
	return p.client
}

// 实现Close方法，关闭客户端
func (p *PostgreSQL) Close() error {
	if p.client == nil {
		return nil
	}
	err := p.client.Close()
	if err == nil {
		log.Println("数据库已关闭")
	}
	return err
}

// 实现Migrate方法，执行数据库迁移
func (p *PostgreSQL) Migrate(ctx context.Context) error {
	if err := p.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("创建数据库Schema失败： %w", err)
	}
	log.Println("数据库迁移完成")
	return nil
}
