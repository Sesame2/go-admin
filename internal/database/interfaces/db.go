package interfaces

import (
	"context"

	"github.com/Sesame2/go-admin/internal/models/ent"
)

type Database interface {
	Client() *ent.Client
	Close() error
	Migrate(ctx context.Context) error
}
