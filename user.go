package otf

import (
	"context"
	"time"

	"github.com/alexedwards/scs/v2"
)

// Workspace represents a user.
type User struct {
	ID string `db:"user_id" jsonapi:"primary,users"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	Username string
	Email    string
}

type Session struct {
	Token  string
	Expiry time.Time
}

type SessionStore interface {
	scs.Store
}

type UserService interface {
	Sessions(ctx context.Context) ([]*Session, error)
}

type UserCreateOptions struct {
	Username string
	Email    string
}

func NewUser(opts UserCreateOptions) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   opts.Username,
		Email:      opts.Email,
	}

	return &user
}
