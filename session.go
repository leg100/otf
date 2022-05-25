package otf

import (
	"context"
	"fmt"
	"time"
)

const (
	DefaultSessionExpiry           = 24 * time.Hour
	FlashSuccessType     FlashType = "success"
	FlashErrorType       FlashType = "error"
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

	// Web app flash message
	Flash *Flash
}

type FlashType string

type Flash struct {
	Type    FlashType
	Message string
}

func FlashSuccess(msg ...interface{}) *Flash {
	return flash(FlashSuccessType, msg...)
}

func FlashError(msg ...interface{}) *Flash {
	return flash(FlashErrorType, msg...)
}

func flash(t FlashType, msg ...interface{}) *Flash {
	return &Flash{
		Type:    t,
		Message: fmt.Sprint(msg...),
	}
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
	// PopFlash reads a flash message from a persistence store before purging
	// it. The token identifies the session.
	PopFlash(ctx context.Context, token string) (*Flash, error)
	// SetFlash writes a flash message the persistence store for the session
	// identified by token.
	SetFlash(ctx context.Context, token string, flash *Flash) error
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
}
