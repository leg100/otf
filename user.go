package otf

import (
	"context"
	"time"

	"github.com/alexedwards/scs/v2"
)

const (
	// Session data keys
	UsernameSessionKey = "username"
	AddressSessionKey  = "ip_address"
	FlashSessionKey    = "flash"
)

// Users represents an oTF user account.
type User struct {
	ID string `db:"user_id" jsonapi:"primary,users"`

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	Username string

	// A user has many sessions
	Sessions []*Session
}

type UserService interface {
	// Login logs a user into oTF. A user is created if they don't already
	// exist. The user is associated with an active session.
	Login(ctx context.Context, opts UserLoginOptions) error

	// Get retrieves a user using their username
	Get(ctx context.Context, username string) (*User, error)

	// Revoke a session belong to user
	RevokeSession(ctx context.Context, token, username string) error
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
	List(ctx context.Context) ([]*User, error)
	LinkSession(ctx context.Context, token, username string) error
	RevokeSession(ctx context.Context, token, username string) error
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

func NewUser(opts UserLoginOptions) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   opts.Username,
	}

	return &user
}

// IsActive queries whether session is the active session. Relies on the
// activeToken being the token for the active session.
func (s *Session) IsActive(activeToken string) bool {
	if s.Token == activeToken {
		return true
	}
	return false
}

func (s *Session) Address() (string, error) {
	data, err := s.decode()
	if err != nil {
		return "", err
	}

	addr, ok := data[AddressSessionKey]
	if !ok {
		return "", nil
	}
	return addr.(string), nil
}

func (s *Session) decode() (map[string]interface{}, error) {
	_, data, err := (scs.GobCodec{}).Decode(s.Data)
	return data, err
}
