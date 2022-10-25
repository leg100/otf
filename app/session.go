package app

import (
	"context"

	"github.com/leg100/otf"
)

// CreateSession creates and persists a user session.
func (a *Application) CreateSession(ctx context.Context, userID, address string) (*otf.Session, error) {
	session, err := otf.NewSession(userID, address)
	if err != nil {
		a.Error(err, "building new session", "uid", userID)
		return nil, err
	}
	if err := a.db.CreateSession(ctx, session); err != nil {
		a.Error(err, "creating session", "uid", userID)
		return nil, err
	}

	a.V(1).Info("created session", "uid", userID)

	return session, nil
}

func (a *Application) GetSessionByToken(ctx context.Context, token string) (*otf.Session, error) {
	return a.db.GetSessionByToken(ctx, token)
}

func (a *Application) ListSessions(ctx context.Context, userID string) ([]*otf.Session, error) {
	return a.db.ListSessions(ctx, userID)
}

func (a *Application) DeleteSession(ctx context.Context, token string) error {
	// Retrieve user purely for logging purposes
	user, err := a.GetUser(ctx, otf.UserSpec{SessionToken: &token})
	if err != nil {
		return err
	}

	if err := a.db.DeleteSession(ctx, token); err != nil {
		a.Error(err, "deleting session", "username", user.Username())
		return err
	}

	a.V(1).Info("deleted session", "username", user.Username())

	return nil
}
