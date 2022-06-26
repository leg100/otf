/*
Package sql implements persistent storage using the sql database.
*/
package sql

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

// String converts a go-string into a postgres non-null string
func String(s string) pgtype.Text {
	return pgtype.Text{String: s, Status: pgtype.Present}
}

// Timestamptz converts a go-time into a postgres non-null timestamptz
func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Status: pgtype.Present}
}

func databaseError(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case noRowsInResultError(err):
		return otf.ErrResourceNotFound
	case errors.As(err, &pgErr):
		switch pgErr.Code {
		case "23505": // unique violation
			return otf.ErrResourcesAlreadyExists
		}
		fallthrough
	default:
		return err
	}
}

func noRowsInResultError(err error) bool {
	for {
		err = errors.Unwrap(err)
		if err == nil {
			return false
		} else if err.Error() == "no rows in result set" {
			return true
		}
	}
}

func includeRelation(includes *string, relation string) bool {
	if includes != nil {
		includes := strings.Split(*includes, ",")
		for _, inc := range includes {
			if inc == relation {
				return true
			}
		}
	}
	return false
}

func includeOrganization(includes *string) bool {
	return includeRelation(includes, "organization")
}

func includeWorkspace(includes *string) bool {
	return includeRelation(includes, "workspace")
}

// conn is a postgres connection, i.e. *pgx.Pool, *pgx.Tx, etc
type conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}
