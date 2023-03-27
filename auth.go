package otf

import (
	"context"
	"net/http"
	"time"
)

const (
	SiteAdminID     = "user-site-admin"
	DefaultUserID   = "user-123"
	DefaultUsername = "otf"
)

type Session interface {
	Expiry() time.Time
	SetCookie(w http.ResponseWriter)
}

type SessionService interface {
	// CreateSession creates a user session.
	CreateSession(r *http.Request, userID string) (Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
}
