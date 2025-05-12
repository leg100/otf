package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leg100/otf/internal"
)

// max conns avail in a pgx pool
const defaultMaxConnections = 10

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
	}

	Connection = genericConnection
)

// New migrates the database to the latest migration version, and then
// constructs and returns a connection pool.
func New(ctx context.Context, logger logr.Logger, connString string) (*DB, error) {
	if err := migrate(ctx, logger, connString); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	// Bump max number of connections in a pool. By default pgx sets it to the
	// greater of 4 or the num of CPUs. However, otfd acquires several dedicated
	// connections for session-level advisory locks and can easily exhaust this.
	connString, err := setDefaultMaxConnections(connString, defaultMaxConnections)
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
		// Set location to UTC for times scanned from database. This ensures
		// that tests for equality pass.
		//
		// See: https://github.com/jackc/pgx/issues/1945#issuecomment-2002077247
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
		})
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("connected to database", "connstr", connString)

	return &DB{Pool: pool, Logger: logger}, nil
}

func (db *DB) Conn(ctx context.Context) Connection {
	if conn, ok := fromContext(ctx); ok {
		return conn
	}
	return db.Pool
}

func (db *DB) Query(ctx context.Context, sql string, args ...any) pgx.Rows {
	rows, _ := db.Conn(ctx).Query(ctx, sql, args...)
	return rows
}

// queryRowResult wraps the error returned by pgx.Row.Scan()
type queryRowResult struct {
	pgx.Row
}

func (r *queryRowResult) Scan(dest ...any) error {
	if err := r.Row.Scan(dest...); err != nil {
		return toError(err)
	}
	return nil
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...any) *queryRowResult {
	row := db.Conn(ctx).QueryRow(ctx, sql, args...)
	return &queryRowResult{Row: row}
}

// Exec executes the sql with the given args. It assumes the command is a row
// affecting command and returns an error if the command does not affect any
// rows.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	cmdTag, err := db.Conn(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return pgconn.CommandTag{}, toError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, internal.ErrResourceNotFound
	}
	return cmdTag, nil
}

func (db *DB) ExecAndPublishEvent(ctx context.Context, obj any, sql string, args ...any) (pgconn.CommandTag, error) {
	var action string
	switch {
	case strings.HasPrefix(sql, InsertAction):
		action = InsertAction
	case strings.HasPrefix(sql, UpdateAction):
		action = UpdateAction
	case strings.HasPrefix(sql, DeleteAction):
		action = DeleteAction
	default:
		return pgconn.CommandTag{}, errors.New("unknown action")
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	var cmd pgconn.CommandTag
	err = db.Tx(ctx, func(ctx context.Context) error {
		cmd, err = db.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}
		_, err = db.Conn(ctx).Exec(ctx, `
	INSERT INTO events (
		type,
		action,
		payload
	) VALUES (
		@type,
		@action,
		@payload
	)`, pgx.NamedArgs{
			"type":    reflect.TypeOf(obj).String(),
			"action":  action,
			"payload": payload,
		})
		return err
	})
	return cmd, err
}

// Int is a convenience wrapper for executing a query that returns a single
// integer.
func (db *DB) Int(ctx context.Context, sql string, args ...any) (int64, error) {
	rows := db.Query(ctx, sql, args...)
	return CollectOneRow(rows, pgx.RowTo[int64])
}

// Tx provides the caller with a callback in which all operations are conducted
// within a transaction.
func (db *DB) Tx(ctx context.Context, callback func(context.Context) error) error {
	var conn Connection = db.Pool

	// Use connection from context if found
	if ctxConn, ok := fromContext(ctx); ok {
		conn = ctxConn
	}

	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		ctx = newContext(ctx, tx)
		return callback(ctx)
	})
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

func (db *DB) Lock(ctx context.Context, table string, fn func(context.Context, Connection) error) error {
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
		return fn(ctx, tx)
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
