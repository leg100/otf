package auth

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type registrySessionService interface {
	CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error)
	// GetRegistrySession retrieves a registry session using a token. Intended
	// as means of checking whether a given token is valid.
	GetRegistrySession(ctx context.Context, token string) (otf.RegistrySession, error)

	createRegistrySession(ctx context.Context, organization string) (*registrySession, error)
}

// Registry session services

func (a *Service) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	return a.createRegistrySession(ctx, organization)
}

func (a *Service) GetRegistrySession(ctx context.Context, token string) (otf.RegistrySession, error) {
	session, err := a.db.getRegistrySession(ctx, token)
	if err != nil {
		a.Error(err, "retrieving registry session", "token", "*****")
		return nil, err
	}

	a.V(2).Info("retrieved registry session", "session", session)

	return session, nil
}

func (a *Service) createRegistrySession(ctx context.Context, organization string) (*registrySession, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateRegistrySessionAction, organization)
	if err != nil {
		return nil, err
	}

	session, err := newRegistrySession(organization)
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
