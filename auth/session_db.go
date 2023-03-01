package auth

import (
	"context"

	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

func (db *pgdb) createSession(ctx context.Context, session *Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     sql.String(session.Token()),
		Address:   sql.String(session.Address()),
		Expiry:    sql.Timestamptz(session.Expiry()),
		UserID:    sql.String(session.UserID()),
		CreatedAt: sql.Timestamptz(session.CreatedAt()),
	})
	return err
}

func (db *pgdb) getSessionByToken(ctx context.Context, token string) (*Session, error) {
	result, err := db.FindSessionByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return sessionRow(result).toSession(), nil
}

func (db *pgdb) listSessions(ctx context.Context, userID string) ([]*Session, error) {
	result, err := db.FindSessionsByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, row := range result {
		sessions = append(sessions, sessionRow(row).toSession())
	}
	return sessions, nil
}

func (db *pgdb) deleteSession(ctx context.Context, token string) error {
	_, err := db.DeleteSessionByToken(ctx, sql.String(token))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
