package sql

import (
	"embed"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	mu sync.Mutex

	//go:embed migrations/*.sql
	migrations embed.FS
)

func migrate(logger logr.Logger, connStr string) error {
	mu.Lock()
	defer mu.Unlock()

	goose.SetLogger(&gooseLogger{logger})

	goose.SetBaseFS(migrations)

	goose.SetDialect("pgx")

	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		return fmt.Errorf("connecting to db for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting postgres dialect for migrations: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	return nil
}
