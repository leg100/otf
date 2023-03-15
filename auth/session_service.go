package auth

import (
	"context"
	"net/http"
)

type sessionService interface {
	CreateSession(r *http.Request, userID string) (*Session, error)
	GetSession(ctx context.Context, token string) (*Session, error)

	createSession(r *http.Request, userID string) (*Session, error)
	listSessions(ctx context.Context, userID string) ([]*Session, error)
	deleteSession(ctx context.Context, token string) error
}

func (a *service) CreateSession(r *http.Request, userID string) (*Session, error) {
	return a.createSession(r, userID)
}

func (a *service) GetSession(ctx context.Context, token string) (*Session, error) {
	return a.db.getSessionByToken(ctx, token)
}

func (a *service) DeleteSession(ctx context.Context, token string) error {
	return a.deleteSession(ctx, token)
}

func (a *service) createSession(r *http.Request, userID string) (*Session, error) {
	session, err := newSession(r, userID)
	if err != nil {
		a.Error(err, "building new session", "uid", userID)
		return nil, err
	}
	if err := a.db.createSession(r.Context(), session); err != nil {
		a.Error(err, "creating session", "uid", userID)
		return nil, err
	}

	a.V(2).Info("created session", "uid", userID)

	return session, nil
}

func (a *service) listSessions(ctx context.Context, userID string) ([]*Session, error) {
	return a.db.listSessions(ctx, userID)
}

func (a *service) deleteSession(ctx context.Context, token string) error {
	if err := a.db.deleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session")
		return err
	}

	a.V(2).Info("deleted session")

	return nil
}
