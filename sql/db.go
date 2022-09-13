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
)

// DB provides access to the postgres db
type DB struct {
	conn
	pggen.Querier
}

// Close closes all connections in the pool. Has no effect if connection is not
// a pool.
func (db *DB) Close() {
	pool, ok := db.conn.(*pgxpool.Pool)
	if ok {
		pool.Close()
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
		return err
	}
	if err := callback(newDB(tx)); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		// return original callback error if rollback succeeds
		return err
	}
	return tx.Commit(ctx)
}

// tx is the same as exported Tx but for use within the sql pkg, passing the
// full *DB to the callback.
func (db *DB) tx(ctx context.Context, callback func(tx *DB) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := callback(newDB(tx)); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		// return original callback error if rollback succeeds
		return err
	}
	return tx.Commit(ctx)
}

// New constructs a new DB
//
// TODO: pass in context
func New(logger logr.Logger, path string, cache otf.Cache, cleanupInterval time.Duration) (*DB, error) {
	conn, err := pgxpool.Connect(context.Background(), path)
	if err != nil {
		return nil, err
	}

	if err := migrate(logger, path); err != nil {
		return nil, err
	}

	db := newDB(conn)

	if cleanupInterval > 0 {
		go db.startCleanup(cleanupInterval)
	}

	return db, nil
}

func newDB(conn conn) *DB {
	return &DB{
		conn:    conn,
		Querier: pggen.NewQuerier(conn),
	}
}

// conn is a postgres connection, i.e. *pgx.Pool, *pgx.Tx, etc
type conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}
