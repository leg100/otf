package sql

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leg100/otf/internal"
)

// toError converts the sql error into a domain error.
func toError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return internal.ErrResourceNotFound
	case errors.As(err, &pgErr):
		switch pgErr.Code {
		case "23503": // foreign key violation
			return &internal.ForeignKeyError{PgError: pgErr}
		case "23505": // unique violation
			return internal.ErrResourceAlreadyExists
		}
	}
	return err
}
