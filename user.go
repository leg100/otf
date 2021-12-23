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

	// A user has many sessions
	Sessions []*Session

	// User belongs to an organization
	Organization *Organization `db:"organizations"`
}

type UserService interface {
	// Login logs a user into oTF. A user is created if they don't already
	// exist. The user is associated with an active session.
	Login(ctx context.Context, opts UserLoginOptions) error

	// Get retrieves a user using their username
	Get(ctx context.Context, username string) (*User, error)
}

// UserLoginOptions are the options for logging a user into the system.
type UserLoginOptions struct {
	// Username of the user.
	Username string

	// SessionToken is the token for the active session of the user.
	SessionToken string
}

type UserStore interface {
	Create(ctx context.Context, user *User) error
	List(ctx context.Context, organizationID string) ([]*User, error)
	LinkSession(ctx context.Context, token, username string) error
	Get(ctx context.Context, username string) (*User, error)
	Delete(ctx context.Context, user_id string) error
}

type Session struct {
	Token  string
	Expiry time.Time
	Data   []byte

	// Session belongs to a user
	UserID string
}

type SessionStore interface {
	// Link links the session with a user, using a session token and user_id.
	Link(ctx context.Context, token, user_id string) error
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
