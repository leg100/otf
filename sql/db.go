package sql

import (
	"context"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

const defaultMaxConnections = "20" // max conns avail in a pgx pool

type (
	// DB provides access to the postgres db as well as queries generated from
	// SQL
	DB struct {
		*pgxpool.Pool // db connection pool
		pggen.Querier // generated queries
	}

	// Options for constructing a DB
	Options struct {
		Logger     logr.Logger
		ConnString string
	}
)

// New constructs a new DB
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
	opts.Logger.Info("connected to database", "connstring", connString)

	if err := migrate(opts.Logger, opts.ConnString); err != nil {
		return nil, err
	}

	db := &DB{
		Pool:    pool,
		Querier: pggen.NewQuerier(pool),
	}

	return db, nil
}

func (db *DB) WaitAndLock(ctx context.Context, id int64) (otf.DatabaseLock, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", id)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(otf.DB) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := callback(&DB{db.Pool, pggen.NewQuerier(tx)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func setDefaultMaxConnections(connString string) (string, error) {
	u, err := url.Parse(connString)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Add("pool_max_conns", defaultMaxConnections)
	u.RawQuery = q.Encode()
	return url.PathUnescape(u.String())
}
