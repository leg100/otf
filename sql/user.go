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
func (db UserDB) Get(ctx context.Context, username string) (*otf.User, error) {
	selectBuilder := psql.Select("*").From("users").Where("username = ?", username)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building SQL query: %w", err)
	}

	var user otf.User
	if err := db.DB.Get(&user, sql, args...); err != nil {
		return nil, databaseError(err, sql)
	}

	user.Sessions, err = listSessions(ctx, db.DB, user.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// LinkSession links a session record to a user.
func (db UserDB) LinkSession(ctx context.Context, token, user_id string) error {
	updateBuilder := psql.Update("sessions").
		Set("user_id", user_id).
		Where("token = ?", token)

	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// Delete deletes a user from the DB.
func (db UserDB) Delete(ctx context.Context, user_id string) error {
	_, err := db.Exec("DELETE FROM users WHERE user_id = $1", user_id)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return nil
}

// listSessions lists sessions belonging to the user with the given user_id.
func listSessions(ctx context.Context, db Getter, user_id string) ([]*otf.Session, error) {
	selectBuilder := psql.Select("token, expiry").From("sessions").Where("user_id = ?", user_id)

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
