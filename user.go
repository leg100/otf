package otf

import (
	"context"
	"time"
)

// Users represents an oTF user account.
type User struct {
	ID string `db:"user_id" jsonapi:"primary,users"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	Username string
}

type UserService interface {
	// Login logs a user into oTF. A user is created if they don't already
	// exist. The user is associated with an active session.
	Login(ctx context.Context, opts UserLoginOptions) error

	// Sessions lists sessions belonging to the user.
	Sessions(ctx context.Context, username string) ([]*Session, error)
}

// UserLoginOptions are the options for logging a user into the system.
type UserLoginOptions struct {
	// Username of the user.
	Username string

	// SessionToken is the token for the active session of the user.
	SessionToken string
}

type UserStore interface {
	Create(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, token string, fn func(*Session) error) (*Session, error)
	Get(ctx context.Context, username string) (*User, error)
	Delete(ctx context.Context, token string) error
}

type Session struct {
	Token  string
	Expiry time.Time

	// Session belongs to a user
	User *User `db:"users"`
}

type SessionStore interface {
	Update(ctx context.Context, token string, fn func(*Session) error) (*Session, error)
	Delete(ctx context.Context, token string) error
}

func NewUser(opts UserLoginOptions) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   opts.Username,
	}

	return &user
}
