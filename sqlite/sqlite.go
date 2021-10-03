/*
Package sqlite implements persistent storage using the sqlite database.
*/
package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	gormzerolog "github.com/leg100/gorm-zerolog"
	"github.com/leg100/otf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

type metadata struct {
	ID        uint
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	ExternalID string `db:"external_id"`
}

type StructScannable interface {
	StructScan(dest interface{}) error
}

type Option func(*gorm.Config)

type Logger struct {
	logr.Logger
}

func WithZeroLogger(logger *zerolog.Logger) Option {
	return func(cfg *gorm.Config) {
		cfg.Logger = &gormzerolog.Logger{Zlog: *logger}
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *Logger) Verbose() bool { return true }

func New(logger logr.Logger, path string, opts ...Option) (*sql.DB, error) {
	cfg := &gorm.Config{}
	for _, o := range opts {
		o(cfg)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Avoid "database is locked" errors:
	// https://github.com/mattn/go-sqlite3/issues/274
	db.SetMaxOpenConns(1)

	// Enable WAL. SQLite performs better with the WAL because it allows
	// multiple readers to operate while data is being written.
	db.Exec(`PRAGMA journal_mode = wal;`)

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		logger.Error(err, "creating migrations")
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("sqlite3://%s", path))
	if err != nil {
		logger.Error(err, "new source instance")
	}

	m.Log = &Logger{Logger: logger}

	err = m.Up()
	if err != nil {
		logger.Error(err, "running migrations")
	}

	return db, nil
}

// Gorm scopes: https://gorm.io/docs/advanced_query.html#Scopes

func paginate(opts otf.ListOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		otf.SanitizeListOptions(&opts)

		offset := (opts.PageNumber - 1) * opts.PageSize

		return db.Offset(offset).Limit(opts.PageSize)
	}
}

// setIfChanged sets a key on a map if a != b. Key is set to the value of b.
func setIfChanged(a, b interface{}, m map[string]interface{}, k string) {
	if reflect.DeepEqual(a, b) {
		return
	}
	m[k] = b
}
