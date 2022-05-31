package sql

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.SessionStore = (*SessionDB)(nil)

	DefaultSessionCleanupInterval = 5 * time.Minute
)

type SessionDB struct {
	*pgxpool.Pool
}

func NewSessionDB(conn *pgxpool.Pool, cleanupInterval time.Duration) *SessionDB {
	db := &SessionDB{
		Pool: conn,
	}
	if cleanupInterval > 0 {
		go db.startCleanup(cleanupInterval)
	}
	return db
}

// CreateSession inserts the session, associating it with the user.
func (db SessionDB) CreateSession(ctx context.Context, session *otf.Session) error {
	q := pggen.NewQuerier(db.Pool)
	_, err := q.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     pgtype.Text{String: session.Token, Status: pgtype.Present},
		Address:   pgtype.Text{String: session.Address, Status: pgtype.Present},
		Expiry:    session.Expiry,
		UserID:    pgtype.Text{String: session.UserID, Status: pgtype.Present},
		CreatedAt: session.CreatedAt(),
	})
	return err
}

// TransferSession updates a session row in the sessions table with the given
// session.  The token identifies the session row to update.
func (db SessionDB) TransferSession(ctx context.Context, token, to string) error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.UpdateSessionUserID(ctx,
		pgtype.Text{String: to, Status: pgtype.Present},
		pgtype.Text{String: token, Status: pgtype.Present},
	)
	return err
}

// DeleteSession deletes a user's session from the DB.
func (db SessionDB) DeleteSession(ctx context.Context, token string) error {
	q := pggen.NewQuerier(db.Pool)

	result, err := q.DeleteSessionByToken(ctx, pgtype.Text{String: token, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func (db SessionDB) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		db.deleteExpired()
	}
}

func (db SessionDB) deleteExpired() error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.DeleteSessionsExpired(context.Background())
	return err
}
