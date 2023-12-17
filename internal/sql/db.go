package sql

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pgxv5 "github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/leg100/otf/internal/sql/pggen"
)

// max conns avail in a pgx pool
const defaultMaxConnections = 10

var (
	//go:embed migrations/*
	migrationsDir embed.FS
	migrations    fs.FS
)

func init() {
	// tern expects a fs without the migrations/ parent directory
	m, err := fs.Sub(migrationsDir, "migrations")
	if err != nil {
		panic("could not find embedded database migrations directory")
	}
	migrations = m
}

type (
	// DB provides access to the postgres db as well as queries generated from
	// SQL
	DB struct {
		*pgxpool.Pool // db connection pool
		logr.Logger
	}

	// Options for constructing a DB
	Options struct {
		Logger     logr.Logger
		ConnString string
	}

	genericConnection interface {
		BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
		SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	}
)

// New constructs a new DB connection pool after migrating the schema to the
// latest version.
func New(ctx context.Context, opts Options) (*DB, error) {
	// Migrate database. Tern is used for migrations, and uses pgx v5, whereas
	// pgx v4 is used elsewhere, mainly because pggen is still on v4:
	//
	// https://github.com/jschaf/pggen/issues/74
	//
	// Therefore a v5 connection is established purely for migrations
	migrateConn, err := pgxv5.Connect(ctx, opts.ConnString)
	if err != nil {
		return nil, fmt.Errorf("connecting to database for migrations: %w", err)
	}
	migrator, err := tern.NewMigrator(ctx, migrateConn, "public.schema_versions")
	if err != nil {
		return nil, fmt.Errorf("constructing database migrator: %w", err)
	}
	if err := migrator.LoadMigrations(migrations); err != nil {
		return nil, fmt.Errorf("loading database migrations: %w", err)
	}
	migrator.OnStart = func(seq int32, name, _, _ string) {
		opts.Logger.V(0).Info("migrating database", "sequence", seq, "name", name)
	}
	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}
	// pgx v4 is used hereon in

	// Bump max number of connections in a pool. By default pgx sets it to the
	// greater of 4 or the num of CPUs. However, otfd acquires several dedicated
	// connections for session-level advisory locks and can easily exhaust this.
	connString, err := setDefaultMaxConnections(opts.ConnString, defaultMaxConnections)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	opts.Logger.Info("connected to database", "connstr", connString)

	return &DB{
		Pool:   pool,
		Logger: opts.Logger,
	}, nil
}

// Conn provides pre-generated queries
func (db *DB) Conn(ctx context.Context) *pggen.DBQuerier {
	if conn, ok := fromContext(ctx); ok {
		return pggen.NewQuerier(conn)
	}
	return pggen.NewQuerier(db.Pool)
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(context.Context, pggen.Querier) error) error {
	var conn genericConnection = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return conn.BeginFunc(ctx, func(tx pgx.Tx) error {
		ctx = newContext(ctx, tx)
		return callback(ctx, pggen.NewQuerier(tx))
	})
}

// Exec acquires a connection from the pool and executes the given SQL. If the
// context contains a transaction then that is used.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if conn, ok := fromContext(ctx); ok {
		return conn.Exec(ctx, sql, args...)
	}
	return db.Pool.Exec(ctx, sql, args...)
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn, ok := fromContext(ctx); ok {
		return conn.QueryRow(ctx, sql, args...)
	}
	return db.Pool.QueryRow(ctx, sql, args...)
}

func (db *DB) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if conn, ok := fromContext(ctx); ok {
		return conn.SendBatch(ctx, b)
	}
	return db.Pool.SendBatch(ctx, b)
}

// WaitAndLock obtains an exclusive session-level advisory lock. If another
// session holds the lock with the given id then it'll wait until the other
// session releases the lock. The given fn is called once the lock is obtained
// and when the fn finishes the lock is released.
func (db *DB) WaitAndLock(ctx context.Context, id int64, fn func(context.Context) error) (err error) {
	// A dedicated connection is obtained. Using a connection pool would cause
	// problems because a lock must be released on the same connection on which
	// it was obtained.
	return db.Pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
		if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", id); err != nil {
			return err
		}
		defer func() {
			_, closeErr := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", id)
			if err != nil {
				db.Error(err, "unlocking session-level advisory lock")
				return
			}
			err = closeErr
		}()
		ctx = newContext(ctx, conn)
		return fn(ctx)
	})
}

func (db *DB) Lock(ctx context.Context, table string, fn func(context.Context, pggen.Querier) error) error {
	var conn genericConnection = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return conn.BeginFunc(ctx, func(tx pgx.Tx) error {
		ctx = newContext(ctx, tx)
		sql := fmt.Sprintf("LOCK TABLE %s IN EXCLUSIVE MODE", table)
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}
		return fn(ctx, pggen.NewQuerier(tx))
	})
}

func setDefaultMaxConnections(connString string, max int) (string, error) {
	// pg connection string can be either a URL or a DSN
	if strings.HasPrefix(connString, "postgres://") || strings.HasPrefix(connString, "postgresql://") {
		u, err := url.Parse(connString)
		if err != nil {
			return "", fmt.Errorf("parsing connection string url: %w", err)
		}
		q := u.Query()
		q.Add("pool_max_conns", strconv.Itoa(max))
		u.RawQuery = q.Encode()
		return url.PathUnescape(u.String())
	} else if connString == "" {
		// presume empty DSN
		return fmt.Sprintf("pool_max_conns=%d", max), nil
	} else {
		// presume non-empty DSN
		return fmt.Sprintf("%s pool_max_conns=%d", connString, max), nil
	}
}
