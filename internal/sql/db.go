package sql

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leg100/otf/internal/sql/sqlc"
)

const (
	// max conns avail in a pgx pool
	defaultMaxConnections = 10
)

type (
	// DB provides access to the postgres db as well as queries generated from
	// SQL
	DB struct {
		*pgxpool.Pool // db connection pool
		logr.Logger
	}

	genericConnection interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
		Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
		SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	}
)

// New migrates the database to the latest migration version, and then
// constructs and returns a connection pool.
func New(ctx context.Context, logger logr.Logger, connString string) (*DB, error) {
	// TODO: move connection code into migrate routing
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	if err := migrate(ctx, logger, conn); err != nil {
		return nil, fmt.Errorf("migrating database schema: %w", err)
	}

	// Bump max number of connections in a pool. By default pgx sets it to the
	// greater of 4 or the num of CPUs. However, otfd acquires several dedicated
	// connections for session-level advisory locks and can easily exhaust this.
	connString, err = setDefaultMaxConnections(connString, defaultMaxConnections)
	if err != nil {
		return nil, err
	}

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// Register table types with pgx, so that it can scan nested tables when
	// querying.
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, t := range tableTypes {
			dt, err := conn.LoadType(ctx, t)
			if err != nil {
				return fmt.Errorf("loading postgres type %s: %w", t, err)
			}
			conn.TypeMap().RegisterType(dt)
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("connected to database", "connstr", connString)

	return &DB{Pool: pool, Logger: logger}, nil
}

// Conn provides pre-generated queries
//
// TODO: rename to reflect role
func (db *DB) Conn(ctx context.Context) *sqlc.Queries {
	if conn, ok := fromContext(ctx); ok {
		return sqlc.New(conn)
	}
	return sqlc.New(db.Pool)
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(context.Context, *sqlc.Queries) error) error {
	var conn genericConnection = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		ctx = newContext(ctx, tx)
		return callback(ctx, sqlc.New(tx))
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

func (db *DB) Lock(ctx context.Context, table string, fn func(context.Context, *sqlc.Queries) error) error {
	var conn genericConnection = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		ctx = newContext(ctx, tx)
		sql := fmt.Sprintf("LOCK TABLE %s IN EXCLUSIVE MODE", table)
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}
		return fn(ctx, sqlc.New(tx))
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
