package sql

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
	"github.com/pkg/errors"
)

// DB provides access to the postgres db
type DB struct {
	conn
	pggen.Querier
	otf.Unmarshaler
}

// Close closes the DB's connections. If the DB has wrapped a transaction then
// this method has no effect.
func (db *DB) Close() {
	switch c := db.conn.(type) {
	case *pgxpool.Conn:
		c.Conn().Close(context.Background())
	case *pgx.Conn:
		c.Close(context.Background())
	case *pgxpool.Pool:
		c.Close()
	default:
		// *pgx.Tx etc
	}
}

func (db *DB) Pool() *pgxpool.Pool {
	pool, ok := db.conn.(*pgxpool.Pool)
	if ok {
		return pool
	}
	return nil
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(tx otf.DB) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "beginning transaction")
	}
	if err := callback(db.copy(tx)); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return errors.Wrap(err, "rolling back transaction")
		}
		// return original callback error if rollback succeeds
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "committing transaction")
	}
	return nil
}

func (db *DB) WaitAndLock(ctx context.Context, id int64, cb func(otf.DB) error) error {
	conn, err := db.Pool().Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "SELECT pg_advisory_lock($1)", id)
	if err != nil {
		return err
	}
	return cb(db.copy(conn))
}

// tx is the same as exported Tx but for use within the sql pkg, passing the
// full *DB to the callback.
func (db *DB) tx(ctx context.Context, callback func(tx *DB) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := callback(db.copy(tx)); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		// return original callback error if rollback succeeds
		return err
	}
	return tx.Commit(ctx)
}

// copy makes a copy of the DB object but with a new connection.
func (db *DB) copy(conn conn) *DB {
	return &DB{
		conn:        conn,
		Querier:     pggen.NewQuerier(conn),
		Unmarshaler: db.Unmarshaler,
	}
}

// New constructs a new DB
func New(ctx context.Context, opts Options) (*DB, error) {
	logger := opts.Logger.WithValues("component", "database")

	conn, err := pgxpool.Connect(ctx, opts.Path)
	if err != nil {
		return nil, err
	}
	logger.Info("successfully connected", "path", opts.Path)

	if err := migrate(logger, opts.Path); err != nil {
		return nil, err
	}

	db := &DB{
		conn:    conn,
		Querier: pggen.NewQuerier(conn),
		Unmarshaler: otf.Unmarshaler{
			CloudService: opts.CloudService,
		},
	}

	if opts.CleanupInterval > 0 {
		go db.startCleanup(ctx, opts.CleanupInterval)
	}

	return db, nil
}

// Options for constructing a DB
type Options struct {
	Logger          logr.Logger
	Path            string
	Cache           otf.Cache
	CleanupInterval time.Duration
	CloudService    otf.CloudService
}

// conn is a postgres connection, i.e. *pgx.Pool, *pgx.Tx, etc
type conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}
