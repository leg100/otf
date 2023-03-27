package auth

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// pgdb is a registry session database on postgres
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

func (db *pgdb) startExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if _, err := db.DeleteSessionsExpired(ctx); err != nil {
				db.Error(err, "purging expired user sessions")
			}
			if _, err := db.DeleteExpiredRegistrySessions(ctx); err != nil {
				db.Error(err, "purging expired registry sessions")
			}
		case <-ctx.Done():
			return
		}
	}
}
