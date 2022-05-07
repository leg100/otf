package sql

import (
	"embed"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func migrate(logger logr.Logger, connStr string) error {
	goose.SetLogger(newGooseLogger(logger))

	goose.SetBaseFS(migrations)

	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		return fmt.Errorf("connecting to db for migrations", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting postgres dialect for migrations: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	return nil
}
