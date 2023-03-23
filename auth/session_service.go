package auth

import (
	"context"
)

type sessionService interface {
	CreateSession(ctx context.Context, opts CreateSessionOptions) (*Session, error)
	GetSession(ctx context.Context, token string) (*Session, error)
	ListSessions(ctx context.Context, userID string) ([]*Session, error)
	DeleteSession(ctx context.Context, token string) error
}

func (a *service) CreateSession(ctx context.Context, opts CreateSessionOptions) (*Session, error) {
	session, err := newSession(opts)
	if err != nil {
		a.Error(err, "building new session")
		return nil, err
	}

	if err := a.db.createSession(ctx, session); err != nil {
		a.Error(err, "creating session", "uid", *opts.UserID)
		return nil, err
	}

	a.V(2).Info("created session", "uid", *opts.UserID)

	return session, nil
}

func (a *service) GetSession(ctx context.Context, token string) (*Session, error) {
	return a.db.getSessionByToken(ctx, token)
}

func (a *service) ListSessions(ctx context.Context, userID string) ([]*Session, error) {
	return a.db.listSessions(ctx, userID)
}

func (a *service) DeleteSession(ctx context.Context, token string) error {
	if err := a.db.deleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session")
		return err
	}

	a.V(2).Info("deleted session")

	return nil
}
