package session

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a database of sessions on postgres
type DB struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *DB {
	return &DB{db}
}

// CreateSession inserts the session, associating it with the user.
func (db *DB) CreateSession(ctx context.Context, session *Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     sql.String(session.Token()),
		Address:   sql.String(session.Address()),
		Expiry:    sql.Timestamptz(session.Expiry()),
		UserID:    sql.String(session.UserID()),
		CreatedAt: sql.Timestamptz(session.CreatedAt()),
	})
	return err
}

func (db *DB) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	result, err := db.FindSessionByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toSession(), nil
}

func (db *DB) ListSessions(ctx context.Context, userID string) ([]*Session, error) {
	result, err := db.FindSessionsByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, row := range result {
		sessions = append(sessions, pgRow(row).toSession())
	}
	return sessions, nil
}

// DeleteSession deletes a user's session from the DB.
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.DeleteSessionByToken(ctx, sql.String(token))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *DB) startSessionExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			db.deleteExpired()
		case <-ctx.Done():
			return
		}
	}
}

func (db *DB) deleteExpired() error {
	_, err := db.DeleteSessionsExpired(context.Background())
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

type pgRow struct {
	Token     pgtype.Text        `json:"token"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	Address   pgtype.Text        `json:"address"`
	Expiry    pgtype.Timestamptz `json:"expiry"`
	UserID    pgtype.Text        `json:"user_id"`
}

func (result pgRow) toSession() *Session {
	return &Session{
		token:     result.Token.String,
		createdAt: result.CreatedAt.Time.UTC(),
		expiry:    result.Expiry.Time.UTC(),
		userID:    result.UserID.String,
		address:   result.Address.String,
	}
}
