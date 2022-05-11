package otf

import "context"

// SessionStore is a persistence store for user sessions.
type SessionStore interface {
	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error

	// UpdateSession persists any updates to a user's session
	UpdateSession(ctx context.Context, token string, updated *Session) error

	// TransferSession transfers a session from one user to another
	TransferSession(ctx context.Context, token string, fromID, toID string) error

	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error
}
