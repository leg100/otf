package sql

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

const defaultMaxConnections = "20" // max conns avail in a pgx pool

type (
	// DB provides access to the postgres db as well as queries generated from
	// SQL
	DB struct {
		*pgxpool.Pool         // db connection pool
		pggen.Querier         // generated queries
		conn          conn    // current connection
		tx            *pgx.Tx // current transaction

		logr.Logger
	}

	// Options for constructing a DB
	Options struct {
		Logger     logr.Logger
		ConnString string
	}

	// conn abstracts a postgres connection, which could be a single connection,
	// a pool of connections, or a transaction.
	conn interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	}
)

// New constructs a new DB connection pool, and migrates the schema to the
// latest version.
func New(ctx context.Context, opts Options) (*DB, error) {
	// Bump max number of connections in a pool. By default pgx sets it to the
	// greater of 4 or the num of CPUs. However, otfd acquires several dedicated
	// connections for session-level advisory locks and can easily exhaust this.
	connString, err := setDefaultMaxConnections(opts.ConnString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	opts.Logger.Info("connected to database", "connstr", connString)

	// goose gets upset with max_pool_conns parameter so pass it the unaltered
	// connection string
	if err := migrate(opts.Logger, opts.ConnString); err != nil {
		return nil, err
	}

	return &DB{
		Pool:    pool,
		Querier: pggen.NewQuerier(pool),
		conn:    pool,
		Logger:  opts.Logger,
	}, nil
}

func (db *DB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return db.conn.Exec(ctx, sql, arguments...)
}

// WaitAndLock obtains an exclusive session-level advisory lock. If another
// session holds the lock with the given id then it'll wait until the other
// session releases the lock. The given fn is called once the lock is obtained
// and when the fn finishes the lock is released.
func (db *DB) WaitAndLock(ctx context.Context, id int64, fn func() error) (err error) {
	// A dedicated connection is obtained. Using a connection pool would cause
	// problems because a lock must be released on the same connection on which
	// it was obtained.
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", id); err != nil {
		return err
	}
	defer func() {
		_, closeErr := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", id)
		if err != nil {
			db.Error(err, "unlocking session lock")
			return
		}
		err = closeErr
	}()

	if err = fn(); err != nil {
		return err
	}
	return
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(otf.DB) error) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := callback(&DB{db.Pool, pggen.NewQuerier(tx), tx, &tx, db.Logger}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func setDefaultMaxConnections(connString string) (string, error) {
	// pg connection string can be either a URL or a DSN
	if strings.HasPrefix(connString, "postgres://") || strings.HasPrefix(connString, "postgresql://") {
		u, err := url.Parse(connString)
		if err != nil {
			return "", fmt.Errorf("parsing connection string url: %w", err)
		}
		q := u.Query()
		q.Add("pool_max_conns", defaultMaxConnections)
		u.RawQuery = q.Encode()
		return url.PathUnescape(u.String())
	} else if connString == "" {
		// presume empty DSN
		return fmt.Sprintf("pool_max_conns=%s", defaultMaxConnections), nil
	} else {
		// presume non-empty DSN
		return fmt.Sprintf("%s pool_max_conns=%s", connString, defaultMaxConnections), nil
	}
}
