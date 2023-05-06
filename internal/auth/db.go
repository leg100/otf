package auth

import (
	"context"

	"github.com/go-logr/logr"
	internal "github.com/leg100/otf"
)

// pgdb stores authentication resources in a postgres database
type pgdb struct {
	internal.DB // provides access to generated SQL queries
	logr.Logger
}

func newDB(database internal.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx internal.DB) error {
		return callback(newDB(tx, db.Logger))
	})
}
