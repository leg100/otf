package otf

import "context"

const (
	// path cookie stores the last path the user attempted to access
	PathCookie = "path"
)

type SessionService interface {
	// CreateSession creates a user session.
	CreateSession(ctx context.Context, userID, address string) (Session, error)
	// GetSession retrieves a session using its token.
	GetSessionByToken(ctx context.Context, token string) (Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
}

type CreateSessionOptions struct{}
