/*
Package sqlite implements persistent storage using the sqlite database.
*/
package sqlite

import (
	"embed"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/leg100/otf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

//go:embed migrations/*.sql
var fs embed.FS

type Getter interface {
	Get(dest interface{}, query string, args ...interface{}) error
}

type metadata struct {
	ID        int64
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	ExternalID string `db:"external_id"`
}

type StructScannable interface {
	StructScan(dest interface{}) error
}

type Option func()

type Logger struct {
	logr.Logger
}

func WithZeroLogger(zlog *zerolog.Logger) Option {
	return func() {
		goose.SetLogger(NewGooseLogger(zlog))
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *Logger) Verbose() bool { return true }

func New(logger logr.Logger, path string, opts ...Option) (*sqlx.DB, error) {
	for _, o := range opts {
		o()
	}

	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Map struct field names from CamelCase to snake_case.
	db.MapperFunc(strcase.ToSnake)

	// Avoid "database is locked" errors:
	// https://github.com/mattn/go-sqlite3/issues/274
	db.SetMaxOpenConns(1)

	// Enable WAL. SQLite performs better with the WAL because it allows
	// multiple readers to operate while data is being written.
	db.Exec(`PRAGMA journal_mode = wal;`)

	goose.SetBaseFS(fs)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("setting sqlite3 dialect for migrations: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return nil, fmt.Errorf("unable to migrate database: %w", err)
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

func FindUpdates(m *reflectx.Mapper, a, b interface{}) map[string]interface{} {
	idx := diffIndex(a, b)
	if len(idx) == 0 {
		return nil
	}

	updates := make(map[string]interface{})

	smap := m.TypeMap(reflect.TypeOf(b))
	fmap := m.FieldMap(reflect.ValueOf(b))
	for _, n := range idx {
		path := smap.GetByTraversal(n).Path
		val := fmap[path].Interface()
		updates[path] = val
	}

	return updates
}

// diffIndex returns an index of differences in the fields of two structs of
// identical types. Supports nested structs.
func diffIndex(a, b interface{}) [][]int {
	return doDiffIndex(reflect.ValueOf(a), reflect.ValueOf(b), nil, nil)
}

func doDiffIndex(v1, v2 reflect.Value, idx [][]int, n []int) [][]int {
	if reflect.DeepEqual(v1.Interface(), v2.Interface()) {
		return idx
	}

	switch v1.Kind() {
	case reflect.Ptr, reflect.Interface:
		idx = doDiffIndex(v1.Elem(), v2.Elem(), idx, n)
	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			idx = doDiffIndex(v1.Field(i), v2.Field(i), idx, append(n, i))
		}
	default:
		idx = append(idx, n)
	}

	return idx
}

// asColumnList takes a table name and a list of columns and returns the SQL
// syntax for a list of column aliases. Toggle prefix to add the table name to
// the alias, separated from the column name with a period, e.g. "t1.c1 AS
// t1.c1".
func asColumnList(table string, prefix bool, cols ...string) (sql string) {
	var asLines []string
	for _, c := range cols {
		if prefix {
			asLines = append(asLines, fmt.Sprintf("%s.%s AS '%[1]s.%s'", table, c))
		} else {
			asLines = append(asLines, fmt.Sprintf("%s.%s AS '%[2]s'", table, c))
		}
	}
	return strings.Join(asLines, ",")
}
