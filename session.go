package otf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgtype"
)

const (
	DefaultSessionExpiry = 24 * time.Hour
)

// Session is a user session
type Session struct {
	Token  string
	Expiry time.Time
	// Name of the Organization the session most recently accessed on the web
	// app.
	SessionData
	createdAt time.Time
	// Session belongs to a user
	UserID string
	// whether session is the active session for a user.
	active bool
}

func NewSession(uid string, data *SessionData) (*Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}
	session := Session{
		createdAt:   CurrentTimestamp(),
		Token:       token,
		SessionData: *data,
		Expiry:      CurrentTimestamp().Add(DefaultSessionExpiry),
		UserID:      uid,
	}
	return &session, nil
}

func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) ID() string           { return s.Token }
func (s *Session) Active() bool         { return s.active }

// SessionData is various session data serialised to the session store as JSON.
type SessionData struct {
	// Client IP address
	Address string
}

func NewSessionData(r *http.Request) (*SessionData, error) {
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	data := SessionData{
		Address: addr,
	}
	return &data, nil
}

type SessionService interface {
	// CreateSession creates a user session.
	CreateSession(ctx context.Context, userID string, data *SessionData) (*Session, error)
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
		Token:     result.Token.String,
		createdAt: result.CreatedAt.Time.UTC(),
		Expiry:    result.Expiry.Time.UTC(),
		UserID:    result.UserID.String,
		SessionData: SessionData{
			Address: result.Address.String,
		},
	}
}
