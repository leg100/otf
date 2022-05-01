package sql

import (
	"embed"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func migrate(logger logr.Logger, conn *pgx.Conn) error {
	goose.SetLogger(newGooseLogger(logger))

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting postgres dialect for migrations: %w", err)
	}

	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	return nil
}
