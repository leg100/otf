package sql

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func migrate(logger logr.Logger, db *sql.DB) error {
	goose.SetLogger(newGooseLogger(logger))

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting postgres dialect for migrations: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	return nil
}
