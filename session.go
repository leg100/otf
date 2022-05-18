package otf

import (
	"fmt"
	"time"
)

// Session is a user session
type Session struct {
	Token  string
	Expiry time.Time
	SessionData

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	// Session belongs to a user
	UserID string
}

func NewSession(uid string, data *SessionData) (*Session, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		Token:       token,
		SessionData: *data,
		Expiry:      time.Now().Add(DefaultSessionExpiry),
		UserID:      uid,
	}

	return &session, nil
}

// SessionData is various session data serialised to the session store as JSON.
type SessionData struct {
	// Client IP address
	Address string

	// Web app flash message
	Flash *Flash
}

const (
	FlashSuccessType FlashType = "success"
	FlashErrorType   FlashType = "error"
)

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
