package otf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/leg100/otf/sql/pggen"
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

// SessionStore is a persistence store for user sessions.
type SessionStore interface {
	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
}

func UnmarshalSessionDBType(typ pggen.Sessions) (*Session, error) {
	session := Session{
		Token:     typ.Token.String,
		createdAt: typ.CreatedAt.Time,
		Expiry:    typ.Expiry.Time,
		UserID:    typ.UserID.String,
		SessionData: SessionData{
			Address: typ.Address.String,
		},
	}
	return &session, nil
}
