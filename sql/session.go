package sql

import (
	"context"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	DefaultSessionCleanupInterval = 5 * time.Minute
)

// CreateSession inserts the session, associating it with the user.
func (db *DB) CreateSession(ctx context.Context, session *otf.Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     String(session.Token),
		Address:   String(session.Address),
		Expiry:    Timestamptz(session.Expiry),
		UserID:    String(session.UserID),
		CreatedAt: Timestamptz(session.CreatedAt()),
	})
	return err
}

// DeleteSession deletes a user's session from the DB.
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.DeleteSessionByToken(ctx, String(token))
	if err != nil {
		return databaseError(err)
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
	if err != nil {
		return databaseError(err)
	}
	return nil
}
