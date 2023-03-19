package auth

import (
	"context"

	"github.com/leg100/otf/rbac"
)

type (
	AgentTokenService interface {
		// GetAgentToken retrieves an agent token using the given token.
		GetAgentToken(ctx context.Context, token string) (*AgentToken, error)
		CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) (*AgentToken, error)

		listAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
		deleteAgentToken(ctx context.Context, id string) (*AgentToken, error)
	}
)

func (a *service) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	at, err := a.db.GetAgentTokenByToken(ctx, token)
	if err != nil {
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(2).Info("retrieved agent token", "organization", at.Organization, "id", at.ID)
	return at, nil
}

func (a *service) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) (*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	token, err := NewAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.CreateAgentToken(ctx, token); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", token.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", token.ID, "subject", subject)
	return token, nil
}

func (a *service) listAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.listAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *service) deleteAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.GetAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, at.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}
