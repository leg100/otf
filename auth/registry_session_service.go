package auth

import (
	"context"

	"github.com/leg100/otf/rbac"
)

type (
	RegistrySessionService interface {
		CreateRegistrySession(ctx context.Context, organization string) (*RegistrySession, error)
		// GetRegistrySession retrieves a registry session using a token. Intended
		// as means of checking whether a given token is valid.
		GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error)

		createRegistrySession(ctx context.Context, organization string) (*RegistrySession, error)
	}
)

// Registry session services

func (a *service) CreateRegistrySession(ctx context.Context, organization string) (*RegistrySession, error) {
	return a.createRegistrySession(ctx, organization)
}

func (a *service) GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error) {
	session, err := a.db.getRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}

func (a *service) createRegistrySession(ctx context.Context, organization string) (*RegistrySession, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateRegistrySessionAction, organization)
	if err != nil {
		return nil, err
	}

	session, err := NewRegistrySession(organization)
	if err != nil {
		a.Error(err, "constructing registry session", "subject", subject, "organization", organization)
		return nil, err
	}
	if err := a.db.createRegistrySession(ctx, session); err != nil {
		a.Error(err, "creating registry session", "subject", subject, "session", session)
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "session", session)

	return session, nil
}
