package otf

import (
	"database/sql/driver"
	"encoding/json"
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
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		Token:       token,
		Timestamps:  NewTimestamps(),
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

	// Current organization
	Organization *string
}

func (sd *SessionData) SetFlash(t FlashType, msg ...interface{}) {
	sd.Flash = &Flash{
		Type:    t,
		Message: fmt.Sprint(msg...),
	}
}

func (sd *SessionData) PopFlash() *Flash {
	ret := sd.Flash
	sd.Flash = nil
	return ret
}

type Flash struct {
	Type    FlashType
	Message string
}

// Value : struct -> db
func (f *Flash) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return json.Marshal(f)
}

// Scan : db -> struct
func (f *Flash) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &f)
}

type FlashType string

const (
	FlashSuccessType FlashType = "success"
	FlashErrorType   FlashType = "error"
)
