package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
)

const (
	defaultExpiry          = 24 * time.Hour
	defaultCleanupInterval = 5 * time.Minute
)

// Session is a user session for the web UI
type Session struct {
	token     string
	expiry    time.Time
	address   string
	createdAt time.Time

	// Session belongs to a user
	userID string
}

// NewSession constructs a new Session
func NewSession(r *http.Request, userID string) (*Session, error) {
	ip, err := otfhttp.GetClientIP(r)
	if err != nil {
		return nil, err
	}

	token, err := otf.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		createdAt: otf.CurrentTimestamp(),
		token:     token,
		address:   ip,
		expiry:    otf.CurrentTimestamp().Add(defaultExpiry),
		userID:    userID,
	}

	return &session, nil
}

func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) ID() string           { return s.token }
func (s *Session) Token() string        { return s.token }
func (s *Session) UserID() string       { return s.userID }
func (s *Session) Address() string      { return s.address }
func (s *Session) Expiry() time.Time    { return s.expiry }

func (s *Session) SetCookie(w http.ResponseWriter) {
	html.SetCookie(w, sessionCookie, s.token, otf.Time(s.Expiry()))
}

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
