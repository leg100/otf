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
	// TransferSession transfers an existing session to a user. The token
	// identifies the session to update. TODO: rename to upgrade/promote,
	// because this only ever used to transfer a session from the anonymous user
	// to a named user.
	TransferSession(ctx context.Context, token, userID string) error
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
}

func UnmarshalSessionDBType(typ pggen.Sessions) (*Session, error) {
	session := Session{
		Token:     typ.Token.String,
		createdAt: typ.CreatedAt,
		Expiry:    typ.Expiry,
		UserID:    typ.UserID.String,
		SessionData: SessionData{
			Address: typ.Address.String,
		},
	}
	return &session, nil
}
