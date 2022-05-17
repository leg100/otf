/*
Package sql implements persistent storage using the sql database.
*/
package sql

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/leg100/otf"
)

func databaseError(err error, sqlstmt string) error {
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
		return fmt.Errorf("running SQL statement: %s resulted in an error: %w", sqlstmt, err)
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
