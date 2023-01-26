package app

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// CreateRegistrySession creates and persists a registry session.
func (a *Application) CreateRegistrySession(ctx context.Context, organization string) (*otf.RegistrySession, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateRegistrySessionAction, organization)
	if err != nil {
		return nil, err
	}

	session, err := otf.NewRegistrySession(organization)
	if err != nil {
		a.Error(err, "constructing registry session", "subject", subject, "organization", organization)
		return nil, err
	}
	if err := a.db.CreateRegistrySession(ctx, session); err != nil {
		a.Error(err, "creating registry session", "subject", subject, "session", session)
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "session", session)

	return session, nil
}

// GetRegistrySession retrieves a registry session using a token. Useful for
// checking token is valid.
func (a *Application) GetRegistrySession(ctx context.Context, token string) (*otf.RegistrySession, error) {
	// No need for authz because caller is providing an auth token.

	session, err := a.db.GetRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}
