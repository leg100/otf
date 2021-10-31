/*
Package sql implements persistent storage using the sql database.
*/
package sql

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/leg100/otf"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var fs embed.FS

// psql is our SQL builder, customized to use postgres placeholders ($N).
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Getter interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
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

	db, err := sqlx.Open("postgres", path)
	if err != nil {
		return db, err
	}

	if err := db.Ping(); err != nil {
		return db, err
	}

	// Map struct field names from CamelCase to snake_case.
	db.MapperFunc(strcase.ToSnake)

	goose.SetBaseFS(fs)

	if err := goose.SetDialect("postgres"); err != nil {
		return db, fmt.Errorf("setting postgres dialect for migrations: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return db, fmt.Errorf("unable to migrate database: %w", err)
	}

	return db, nil
}

// FindUpdates compares two structs of identical type for any differences in
// their struct field values. A mapping is returned: the sqlx db path of the
// field mapped to the value found in the field in struct b. Relations are
// stripped out, i.e. those fields with a period in, e.g. 'parent.child'.
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
		if strings.Contains(path, ".") {
			continue
		}
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

// update performs a SQL UPDATE, setting values for fields that have changed
// between two structs. If the value in after is different from before then it
// is included in the UPDATE. If all fields are identical no UPDATE is
// performed.
func update(mapper *reflectx.Mapper, tx sqlx.Execer, table, idCol string, before, after otf.Updateable) (bool, error) {
	updates := FindUpdates(mapper, before, after)
	if len(updates) == 0 {
		return false, nil
	}

	now := time.Now()
	after.SetUpdatedAt(now)
	updates["updated_at"] = now

	sql := psql.Update(table).Where(fmt.Sprintf("%s = ?", idCol), before.GetID())

	query, args, err := sql.SetMap(updates).ToSql()
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return false, fmt.Errorf("executing SQL statement: %s: %w", query, err)
	}

	return true, nil
}

// asColumnList takes a table name and a list of columns and returns the SQL
// syntax for a list of column aliases. Toggle prefix to add the table name to
// the alias, separated from the column name with a period, e.g. "t1.c1 AS
// t1.c1".
func asColumnList(table string, prefix bool, cols ...string) (sql string) {
	var asLines []string
	for _, c := range cols {
		if prefix {
			asLines = append(asLines, fmt.Sprintf("%s.%s AS \"%[1]s.%s\"", table, c))
		} else {
			asLines = append(asLines, fmt.Sprintf("%s.%s AS \"%[2]s\"", table, c))
		}
	}
	return strings.Join(asLines, ",")
}

func databaseError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		// Swap DB no rows found error for the canonical not found error
		return otf.ErrResourceNotFound
	}
	return err
}
