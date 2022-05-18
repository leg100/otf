package otf

import "context"

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
