package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// pgdb stores authentication resources in a postgres database
type pgdb struct {
	otf.DB // provides access to generated SQL queries
	logr.Logger
}

func newDB(database otf.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newDB(tx, db.Logger))
	})
}
