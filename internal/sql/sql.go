/*
Package sql implements persistent storage using the postgres database.
*/
package sql

import (
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
)

// Bool converts a go-boolean into a postgres non-null boolean
func Bool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Status: pgtype.Present}
}

// BoolPtr converts a go-boolean pointer into a postgres nullable boolean
func BoolPtr(s *bool) pgtype.Bool {
	if s != nil {
		return pgtype.Bool{Bool: *s, Status: pgtype.Present}
	}
	return pgtype.Bool{Status: pgtype.Null}
}

// String converts a go-string into a postgres non-null string
func String(s string) pgtype.Text {
	return pgtype.Text{String: s, Status: pgtype.Present}
}

// StringPtr converts a go-string pointer into a postgres nullable string
func StringPtr(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Status: pgtype.Present}
	}
	return pgtype.Text{Status: pgtype.Null}
}

// Int4 converts a go-int into a postgres non-null int4
func Int4(s int) pgtype.Int4 {
	return pgtype.Int4{Int: int32(s), Status: pgtype.Present}
}

// Int4Ptr converts a go-int pointer into a postgres nullable int4
func Int4Ptr(s *int) pgtype.Int4 {
	if s != nil {
		return pgtype.Int4{Int: int32(*s), Status: pgtype.Present}
	}
	return pgtype.Int4{Status: pgtype.Null}
}

// Int8 converts a go-int into a postgres non-null int8
func Int8(s int) pgtype.Int8 {
	return pgtype.Int8{Int: int64(s), Status: pgtype.Present}
}

// Int8Ptr converts a go-int pointer into a postgres nullable int8
func Int8Ptr(s *int) pgtype.Int8 {
	if s != nil {
		return pgtype.Int8{Int: int64(*s), Status: pgtype.Present}
	}
	return pgtype.Int8{Status: pgtype.Null}
}

// NullString returns a postgres null string
func NullString() pgtype.Text {
	return pgtype.Text{Status: pgtype.Null}
}

// UUID converts a google-go-uuid into a postgres non-null UUID
func UUID(s uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: s, Status: pgtype.Present}
}

// Timestamptz converts a go-time into a postgres non-null timestamptz
func Timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Status: pgtype.Present}
}

// TimestamptzPtr converts a go-time pointer into a postgres nullable timestamptz
func TimestamptzPtr(t *time.Time) pgtype.Timestamptz {
	if t != nil {
		return pgtype.Timestamptz{Time: *t, Status: pgtype.Present}
	}
	return pgtype.Timestamptz{Status: pgtype.Null}
}

// JSON converts a []byte into a postgres JSON type
func JSON(b []byte) pgtype.JSON {
	return pgtype.JSON{Bytes: b, Status: pgtype.Present}
}

// Inet converts net.IP into the postgres type pgtype.Inet
func Inet(ip net.IP) pgtype.Inet {
	mask := net.CIDRMask(32, 0)
	return pgtype.Inet{IPNet: &net.IPNet{Mask: mask, IP: ip}, Status: pgtype.Present}
}

func Error(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case NoRowsInResultError(err):
		return internal.ErrResourceNotFound
	case errors.As(err, &pgErr):
		switch pgErr.Code {
		case "23503": // foreign key violation
			return &internal.ForeignKeyError{PgError: pgErr}
		case "23505": // unique violation
			return internal.ErrResourceAlreadyExists
		}
		fallthrough
	default:
		return err
	}
}

func NoRowsInResultError(err error) bool {
	for {
		err = errors.Unwrap(err)
		if err == nil {
			return false
		} else if err.Error() == "no rows in result set" {
			return true
		}
	}
}
