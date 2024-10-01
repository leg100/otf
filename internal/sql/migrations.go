package sql

import (
	"context"
	"embed"
	"sync"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
)

var (
	mu sync.Mutex

	//go:embed migrations/*.sql
	migrations embed.FS
)

// TODO: move to db.go
func migrate(ctx context.Context, logger logr.Logger, conn *pgx.Conn) error {
	mu.Lock()
	defer mu.Unlock()

	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return err
	}
	if err := m.LoadMigrations(migrations); err != nil {
		return err
	}
	return m.Migrate(ctx)
}
