package app

import (
	"context"

	"github.com/leg100/otf"
)

// CreateRegistrySession creates and persists a registry session.
func (a *Application) CreateRegistrySession(ctx context.Context, organization string) (*otf.RegistrySession, error) {
	session, err := otf.NewRegistrySession(organization)
	if err != nil {
		a.Error(err, "constructing registry session", "organization", organization)
		return nil, err
	}
	if err := a.db.CreateRegistrySession(ctx, session); err != nil {
		a.Error(err, "creating registry session", "session", session)
		return nil, err
	}

	a.V(2).Info("created registry session", "session", session)

	return session, nil
}

// GetRegistrySession retrieves a registry session.
func (a *Application) GetRegistrySession(ctx context.Context, token string) (*otf.RegistrySession, error) {
	session, err := a.db.GetRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}
