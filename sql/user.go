package sql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.UserStore = (*UserDB)(nil)

	insertUserSQL = `INSERT INTO users (user_id, created_at, updated_at, username)
VALUES (:user_id, :created_at, :updated_at, :username)`

	insertSessionSQL = `INSERT INTO sessions (token, data, created_at, updated_at, user_id)
VALUES (:token, :data, :created_at, :updated_at, :user_id)`
)

type UserDB struct {
	*sqlx.DB
}

func NewUserDB(db *sqlx.DB) *UserDB {
	return &UserDB{
		DB: db,
	}
}

// Create persists a User to the DB.
func (db UserDB) Create(ctx context.Context, user *otf.User) error {
	sql, args, err := db.BindNamed(insertUserSQL, user)
	if err != nil {
		return err
	}
	_, err = db.Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (db UserDB) List(ctx context.Context) ([]*otf.User, error) {
	selectBuilder := psql.Select("*").From("users")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var users []*otf.User
	if err := db.Select(&users, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan users from db: %w", err)
	}

	return users, nil
}

// Get retrieves a user from the DB, along with its sessions.
func (db UserDB) Get(ctx context.Context, spec otf.UserSpecifier) (*otf.User, error) {
	selectBuilder := psql.
		Select("*").
		From("users")

	switch {
	case spec.Username != nil:
		selectBuilder = selectBuilder.Where("username = ?", *spec.Username)
	case spec.Token != nil:
		selectBuilder = selectBuilder.Where("token = ?", *spec.Token)
	default:
		return nil, fmt.Errorf("empty user spec provided")
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building SQL query: %w", err)
	}

	// get user
	var user otf.User
	if err := db.DB.Get(&user, sql, args...); err != nil {
		return nil, databaseError(err, sql)
	}

	// ...and their sessions
	user.Sessions, err = listSessions(ctx, db.DB, user.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateSession inserts the session, associating it with the user.
func (db UserDB) CreateSession(ctx context.Context, session *otf.Session) error {
	sql, args, err := db.BindNamed(insertSessionSQL, session)
	if err != nil {
		return err
	}
	_, err = db.Exec(sql, args...)
	if err != nil {
		return err
	}

	var result string
	if err := db.DB.Get(&result, sql, args...); err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// UpdateSession updates a session row in the sessions table with the given
// session. The token identifies the session row to update.
func (db UserDB) UpdateSession(ctx context.Context, token string, updated *otf.Session) error {
	existing, err := getSession(ctx, db.DB, token)
	if err != nil {
		return err
	}

	updateBuilder := psql.
		Update("sessions").
		Where("token = ?", updated.Token)

	var modified bool

	if existing.SessionData != updated.SessionData {
		modified = true
		updateBuilder = updateBuilder.Set("data", updated.SessionData)
	}

	if existing.Expiry != updated.Expiry {
		modified = true
		updateBuilder = updateBuilder.Set("expiry", updated.Expiry)
	}

	if existing.UserID != updated.UserID {
		modified = true
		updateBuilder = updateBuilder.Set("user_id", updated.Expiry)
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

// LinkSession (re-)associates a session with a user.
func (db UserDB) LinkSession(ctx context.Context, session *otf.Session, user *otf.User) error {
	updateBuilder := psql.
		Update("sessions").
		Set("user_id", user.ID).
		Where("token = ?", session.Token)

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

// Delete deletes a user from the DB.
func (db UserDB) Delete(ctx context.Context, spec otf.UserSpecifier) error {
	deleteBuilder := psql.Delete("users")

	switch {
	case spec.Username != nil:
		deleteBuilder = deleteBuilder.Where("username = ?", *spec.Username)
	case spec.Token != nil:
		deleteBuilder = deleteBuilder.Where("token = ?", *spec.Token)
	default:
		return fmt.Errorf("empty user spec provided")
	}

	sql, args, err := deleteBuilder.ToSql()
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
func (db UserDB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("unable to delete session: %w", err)
	}

	return nil
}

// listSessions lists sessions belonging to the user with the given userID.
func listSessions(ctx context.Context, db Getter, userID string) ([]*otf.Session, error) {
	selectBuilder := psql.Select("token, data, expiry").From("sessions").Where("user_id = ?", userID)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var sessions []*otf.Session
	if err := db.Select(&sessions, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan sessions from db: %w", err)
	}

	return sessions, nil
}

func getSession(ctx context.Context, db Getter, token string) (*otf.Session, error) {
	selectBuilder := psql.
		Select("*").
		From("sessions").
		Where("token = ?", token)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building SQL query: %w", err)
	}

	var session otf.Session
	if err := db.Get(&session, sql, args...); err != nil {
		return nil, databaseError(err, sql)
	}

	return &session, nil
}
