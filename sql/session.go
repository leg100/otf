package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.SessionStore = (*SessionDB)(nil)

	DefaultSessionCleanupInterval = 5 * time.Minute

	sessionColumns = []string{
		"token",
		"created_at",
		"updated_at",
		"flash",
		"address",
		"expiry",
		"user_id",
	}

	tokenColumns = []string{
		"token_id",
		"created_at",
		"updated_at",
		"description",
		"user_id",
	}

	insertSessionSQL = `INSERT INTO sessions (token, flash, address, created_at, updated_at, expiry, user_id)
VALUES (:token, :flash, :address, :created_at, :updated_at, :expiry, :user_id)`

	insertTokenSQL = `INSERT INTO tokens (token_id, token, created_at, updated_at, description, user_id)
VALUES (:token_id, :token, :created_at, :updated_at, :description, :user_id)`
)

type userRow interface {
	GetUserID() *string
	GetUsername() *string
	GetCurrentOrganization() *string

	Timestamps
}

type SessionDB struct {
	*pgx.Conn
}

func NewSessionDB(conn *pgx.Conn, cleanupInterval time.Duration) *SessionDB {
	udb := &SessionDB{
		Conn: conn,
	}
	if cleanupInterval > 0 {
		go udb.startCleanup(cleanupInterval)
	}
	return udb
}

// CreateSession inserts the session, associating it with the user.
func (db SessionDB) CreateSession(ctx context.Context, session *otf.Session) error {
	q := NewQuerier(db.Conn)

	_, err := q.InsertSession(ctx, InsertSessionParams{
		Token:   &session.Token,
		Address: &session.Address,
		Expiry:  session.Expiry,
		UserID:  &session.UserID,
	})
	return err
}

// UpdateSession updates a session row in the sessions table with the given
// session. The token identifies the session row to update.
func (db SessionDB) UpdateSession(ctx context.Context, token string, updated *otf.Session) error {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	existing, err := getSession(ctx, q, token)
	if err != nil {
		return err
	}

	updateBuilder := psql.
		Update("sessions").
		Where("token = ?", updated.Token)

	var modified bool

	if existing.Address != updated.Address {
		return fmt.Errorf("address cannot be updated on a session")
	}

	if existing.Flash != updated.Flash {
		modified = true
		updateBuilder = updateBuilder.Set("flash", updated.Flash)
	}

	if existing.Expiry != updated.Expiry {
		modified = true
		updateBuilder = updateBuilder.Set("expiry", updated.Expiry)
	}

	if existing.UserID != updated.UserID {
		modified = true
		updateBuilder = updateBuilder.Set("user_id", updated.UserID)
	}

	if !modified {
		return fmt.Errorf("update was requested but no changes were found")
	}

	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// TransferSession updates a session row in the sessions table with the given
// session.  The token identifies the session row to update.
func (db SessionDB) TransferSession(ctx context.Context, from, to string, updated *otf.Session) error {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	existing, err := getSession(ctx, q, token)
	if err != nil {
		return err
	}

	updateBuilder := psql.
		Update("sessions").
		Where("token = ?", updated.Token)

	var modified bool

	if existing.Address != updated.Address {
		return fmt.Errorf("address cannot be updated on a session")
	}

	if existing.Flash != updated.Flash {
		modified = true
		updateBuilder = updateBuilder.Set("flash", updated.Flash)
	}

	if existing.Expiry != updated.Expiry {
		modified = true
		updateBuilder = updateBuilder.Set("expiry", updated.Expiry)
	}

	if existing.UserID != updated.UserID {
		modified = true
		updateBuilder = updateBuilder.Set("user_id", updated.UserID)
	}

	if !modified {
		return fmt.Errorf("update was requested but no changes were found")
	}

	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// DeleteSession deletes a user's session from the DB.
func (db SessionDB) DeleteSession(ctx context.Context, token string) error {
	q := NewQuerier(db.Conn)

	result, err := q.DeleteSessionByToken(ctx, &token)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
