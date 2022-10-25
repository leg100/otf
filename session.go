package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

const (
	DefaultSessionExpiry = 24 * time.Hour
)

// Session is a user session
type Session struct {
	token     string
	expiry    time.Time
	address   string
	createdAt time.Time

	// Session belongs to a user
	userID string
}

func NewSession(uid, address string) (*Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}
	session := Session{
		createdAt: CurrentTimestamp(),
		token:     token,
		address:   address,
		expiry:    CurrentTimestamp().Add(DefaultSessionExpiry),
		userID:    uid,
	}
	return &session, nil
}

func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) ID() string           { return s.token }
func (s *Session) Token() string        { return s.token }
func (s *Session) UserID() string       { return s.userID }
func (s *Session) Address() string      { return s.address }
func (s *Session) Expiry() time.Time    { return s.expiry }

type NewSessionOption func(*Session)

func SessionExpiry(expiry time.Time) NewSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}

type SessionService interface {
	// CreateSession creates a user session.
	CreateSession(ctx context.Context, userID, address string) (*Session, error)
	// GetSession retrieves a session using its token.
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]*Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
}

// SessionStore is a persistence store for user sessions.
type SessionStore interface {
	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error
	// GetSession retrieves a session using its token.
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]*Session, error)
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
}

type SessionResult struct {
	Token     pgtype.Text        `json:"token"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	Address   pgtype.Text        `json:"address"`
	Expiry    pgtype.Timestamptz `json:"expiry"`
	UserID    pgtype.Text        `json:"user_id"`
}

func UnmarshalSessionResult(result SessionResult) *Session {
	return &Session{
		token:     result.Token.String,
		createdAt: result.CreatedAt.Time.UTC(),
		expiry:    result.Expiry.Time.UTC(),
		userID:    result.UserID.String,
		address:   result.Address.String,
	}
}
