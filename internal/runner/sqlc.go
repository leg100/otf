// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package runner

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func New() *Queries {
	return &Queries{}
}

type Queries struct {
}
