package sql

import (
	"context"
	"encoding/json"
	"time"

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

	result, err := q.InsertSession(ctx, pggen.InsertSessionParams{
		Token:   session.Token,
		Address: session.Address,
		Expiry:  session.Expiry,
		UserID:  session.UserID,
	})
	if err != nil {
		return err
	}
	session.CreatedAt = result.CreatedAt
	session.UpdatedAt = result.UpdatedAt

	return err
}

func (db SessionDB) SetFlash(ctx context.Context, token string, flash *otf.Flash) error {
	q := pggen.NewQuerier(db.Pool)

	data, err := json.Marshal(flash)
	if err != nil {
		return err
	}

	_, err = q.UpdateSessionFlashByToken(ctx, data, token)
	if err != nil {
		return err
	}

	return nil
}

func (db SessionDB) PopFlash(ctx context.Context, token string) (*otf.Flash, error) {
	q := pggen.NewQuerier(db.Pool)

	data, err := q.FindSessionFlashByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// No flash found
	if data == nil {
		return nil, nil
	}

	// Set flash in DB to NULL
	if _, err := q.UpdateSessionFlashByToken(ctx, nil, token); err != nil {
		return nil, err
	}

	// Marshal bytes into flash obj
	var flash otf.Flash
	if err := json.Unmarshal(data, &flash); err != nil {
		return nil, err
	}

	return &flash, nil
}

// TransferSession updates a session row in the sessions table with the given
// session.  The token identifies the session row to update.
func (db SessionDB) TransferSession(ctx context.Context, token, to string) error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.UpdateSessionUserID(ctx, to, token)
	return err
}

// DeleteSession deletes a user's session from the DB.
func (db SessionDB) DeleteSession(ctx context.Context, token string) error {
	q := pggen.NewQuerier(db.Pool)

	result, err := q.DeleteSessionByToken(ctx, token)
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
