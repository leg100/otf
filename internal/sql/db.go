package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
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

	connection interface {
		Begin(ctx context.Context) (pgx.Tx, error)
		Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	}
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

func (db *DB) Query(ctx context.Context, sql string, args ...any) pgx.Rows {
	rows, _ := db.conn(ctx).Query(ctx, sql, args...)
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
	row := db.conn(ctx).QueryRow(ctx, sql, args...)
	return &queryRowResult{Row: row}
}

// Exec executes the sql with the given args. It assumes the command is a row
// affecting command and returns an error if the command does not affect any
// rows.
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	cmdTag, err := db.conn(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return pgconn.CommandTag{}, toError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, internal.ErrResourceNotFound
	}
	return cmdTag, nil
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
	var conn connection = db.Pool

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

// Notify sends a postgres notification. The payload is marshaled as JSON and
// sent to the given channel.
func (db *DB) Notify(ctx context.Context, channel string, payload any) error {
	marshaled, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = db.conn(ctx).Exec(ctx, fmt.Sprintf("NOTIFY %s, '%s'", channel, string(marshaled)))
	if err != nil {
		return fmt.Errorf("sending postgres notification: %w", err)
	}
	return nil
}

// Listen to a postgres notification channel, returning a channel of
// notifications.
func (db *DB) Listen(ctx context.Context, channel string) (<-chan string, error) {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire postgres connection: %w", err)
	}

	if _, err := conn.Exec(ctx, "LISTEN "+channel); err != nil {
		conn.Release()
		return nil, err
	}

	ch := make(chan string)
	go func() {
		defer conn.Release()
		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					// parent has decided to shutdown so exit without logging an error
				default:
					db.Error(err, "waiting for postgres notification")
				}
				close(ch)
				return
			}
			select {
			case <-ctx.Done():
				// parent has decided to shutdown so exit without logging an error
				close(ch)
				return
			case ch <- notification.Payload:
			}
		}
	}()
	return ch, nil
}

func (db *DB) Lock(ctx context.Context, table string, fn func(context.Context) error) error {
	var conn connection = db.Pool

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
		return fn(ctx)
	})
}

func (db *DB) conn(ctx context.Context) connection {
	if conn, ok := fromContext(ctx); ok {
		return conn
	}
	return db.Pool
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
