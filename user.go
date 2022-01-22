package otf

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/alexedwards/scs/v2"
)

const (
	// Session data keys
	UsernameSessionKey = "username"
	AddressSessionKey  = "ip_address"
	FlashSessionKey    = "flash"

	DefaultSessionExpiry = 24 * time.Hour

	AnonymousUsername string = "anonymous"
)

// User represents an oTF user account.
type User struct {
	// ID uniquely identifies users
	ID string `db:"user_id" jsonapi:"primary,users"`

	// Username is the SSO-provided username
	Username string

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	// A user has many sessions
	Sessions []*Session

	// The currently active session. The value is nil if there is no active
	// session.
	ActiveSession *Session
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// NewAnonymousSession creates a new session for the anonymous user.
	NewAnonymousSession(ctx context.Context) (*User, error)

	// Promote promotes an anonymous user to the named user.
	Promote(ctx context.Context, anon *User, username string) (*User, error)

	// Get retrieves a user according to the spec.
	Get(ctx context.Context, spec UserSpecifier) (*User, error)

	// Revoke a session belong to user
	RevokeSession(ctx context.Context, token, username string) error
}

// UserStore is a persistence store for user accounts and their associated
// sessions.
type UserStore interface {
	Create(ctx context.Context, user *User) error
	List(ctx context.Context) ([]*User, error)

	// CreateSession persists session to the store.
	CreateSession(ctx context.Context, session *Session) error

	// LinkSession associates the session with the user.
	LinkSession(ctx context.Context, session *Session, user *User) error

	RevokeSession(ctx context.Context, token, username string) error
	Get(ctx context.Context, spec UserSpecifier) (*User, error)
	Delete(ctx context.Context, userID string) error
}

type UserSpecifier struct {
	Username *string
	Token    *string
}

// KeyValue returns the user specifier in key-value form. Useful for logging
// purposes.
func (spec *UserSpecifier) KeyValue() []interface{} {
	if spec.Username != nil {
		return []interface{}{"username", *spec.Username}
	}
	if spec.Token != nil {
		return []interface{}{"token", *spec.Token}
	}

	return []interface{}{"invalid user spec", ""}
}

// Session is a user session
type Session struct {
	Token  string
	Expiry time.Time
	Data   []byte

	// Session belongs to a user
	UserID string
}

func NewUser(username string) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   username,
	}

	return &user
}

// AttachNewSession creates and attaches a new session to the user. The new
// session is made the active session for the user.
func (u *User) AttachNewSession() (*Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		Token:  token,
		Data:   make([]byte, 0),
		Expiry: time.Now().Add(DefaultSessionExpiry),
		UserID: u.ID,
	}

	u.Sessions = append(u.Sessions, &session)

	u.ActiveSession = &session

	return &session, nil
}

// IsActive queries whether session is the active session. Relies on the
// activeToken being the token for the active session.
func (s *Session) IsActive(activeToken string) bool {
	return s.Token == activeToken
}

// Address gets the source IP address for the user session.
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

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
