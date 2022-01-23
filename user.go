package otf

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
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
}

// AttachNewSession creates and attaches a new session to the user.
func (u *User) AttachNewSession(data SessionData) (*Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		Token:  token,
		Data:   data,
		Expiry: time.Now().Add(DefaultSessionExpiry),
		UserID: u.ID,
	}

	u.Sessions = append(u.Sessions, &session)

	return &session, nil
}

// IsAuthenticated determines if the user is authenticated, i.e. not an
// anonymous user.
func (u *User) IsAuthenticated() bool {
	return u.Username != AnonymousUsername
}

func (u *User) String() string {
	return u.Username
}

// TransferSession transfers a session from the receiver to another user.
func (u *User) TransferSession(session *Session, to *User) {
	// Remove session from receiver
	var receiverSessions []*Session
	for _, s := range u.Sessions {
		if s.Token != session.Token {
			receiverSessions = append(receiverSessions, s)
		}
	}
	u.Sessions = receiverSessions

	// Add session to destination user
	to.Sessions = append(to.Sessions, session)

	// Update session's user reference
	session.UserID = to.ID
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// Create creates a user with the given username.
	Create(ctx context.Context, username string) (*User, error)

	// CreateSession creates a user session.
	CreateSession(ctx context.Context, user *User, data SessionData) (*Session, error)

	// TransferSession transfers a session from one user to another.
	TransferSession(ctx context.Context, token string, from, to *User) (*Session, error)

	// Get retrieves a user according to the spec.
	Get(ctx context.Context, spec UserSpecifier) (*User, error)

	// Get retrieves the anonymous user.
	GetAnonymous(ctx context.Context) (*User, error)

	// UpdateSession persists any updates to the user's session data
	UpdateSessionData(ctx context.Context, token, key string, val interface{}) error

	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
}

// UserStore is a persistence store for user accounts and their associated
// sessions.
type UserStore interface {
	Create(ctx context.Context, user *User) error

	Get(ctx context.Context, spec UserSpecifier) (*User, error)

	List(ctx context.Context) ([]*User, error)

	Delete(ctx context.Context, spec UserSpecifier) error

	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error

	// UpdateSession persists any updates to a user's session
	UpdateSession(ctx context.Context, token string, fn func(*Session) error) (*Session, error)

	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
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
	Data   SessionData

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	// Session belongs to a user
	UserID string
}

func (s *Session) GetID() string  { return s.Token }
func (s *Session) String() string { return s.Token }

// SessionData is arbitrary session data
type SessionData map[string]interface{}

// Value: struct -> db
func (sd SessionData) Value() (driver.Value, error) {
	return json.Marshal(sd)
}

// Scan: db -> struct
func (sd SessionData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &sd)
}

func NewUser(username string) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   username,
	}

	return &user
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
