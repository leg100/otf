package sql

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	DefaultSessionCleanupInterval = 5 * time.Minute
)

// CreateSession inserts the session, associating it with the user.
func (db *DB) CreateSession(ctx context.Context, session *otf.Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     pgtype.Text{String: session.Token, Status: pgtype.Present},
		Address:   pgtype.Text{String: session.Address, Status: pgtype.Present},
		Expiry:    session.Expiry,
		UserID:    pgtype.Text{String: session.UserID, Status: pgtype.Present},
		CreatedAt: session.CreatedAt(),
	})
	return err
}

// DeleteSession deletes a user's session from the DB.
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	result, err := db.DeleteSessionByToken(ctx, pgtype.Text{String: token, Status: pgtype.Present})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}
	return nil
}

func (db *DB) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		db.deleteExpired()
	}
}

func (db *DB) deleteExpired() error {
	_, err := db.DeleteSessionsExpired(context.Background())
	return err
}
