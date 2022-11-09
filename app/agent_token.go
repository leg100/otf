package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateAgentToken(ctx context.Context, opts otf.AgentTokenCreateOptions) (*otf.AgentToken, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateAgentTokenAction, opts.OrganizationName)
	if err != nil {
		return nil, err
	}

	token, err := otf.NewAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.CreateAgentToken(ctx, token); err != nil {
		a.Error(err, "creating agent token", "organization", opts.OrganizationName, "id", token.ID(), "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.OrganizationName, "id", token.ID(), "subject", subject)
	return token, nil
}

func (a *Application) ListAgentTokens(ctx context.Context, organizationName string) ([]*otf.AgentToken, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.ListAgentTokensAction, organizationName)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.ListAgentTokens(ctx, organizationName)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organizationName, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed agent tokens", "organization", organizationName, "subject", subject)
	return tokens, nil
}

// GetAgentToken retrieves the agent token metadata corresponding to the given
// token.
func (a *Application) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	at, err := a.db.GetAgentToken(ctx, token)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(2).Info("retrieved agent token", "organization", at.OrganizationName(), "id", at.ID())
	return at, nil
}

func (a *Application) DeleteAgentToken(ctx context.Context, id string, organizationName string) error {
	subject, err := a.CanAccessOrganization(ctx, otf.DeleteAgentTokenAction, organizationName)
	if err != nil {
		return err
	}

	if err := a.db.DeleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "id", id, "subject", subject)
		return err
	}
	a.V(0).Info("deleted agent token", "id", id, "subject", subject)
	return nil
}
