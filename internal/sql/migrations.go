package sql

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/leg100/otf/internal/logr"
	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

func migrate(ctx context.Context, logger logr.Logger, connString string) error {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator: %w", err)
	}
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}
	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations: %w", err)
	}
	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retreiving current database migration version")
	}
	if err := m.Migrate(ctx); err != nil {
		return err
	}
	if from == int32(len(m.Migrations)) {
		logger.Info("database schema up to date", "version", len(m.Migrations))
	} else {
		logger.Info("migrated database schema", "from", from, "to", len(m.Migrations))
	}
	return nil
}
