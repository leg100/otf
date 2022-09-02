package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateAgentToken(ctx context.Context, opts otf.AgentTokenCreateOptions) (*otf.AgentToken, error) {
	token, err := otf.NewAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.CreateAgentToken(ctx, token); err != nil {
		a.Error(err, "creating agent token", "organization", opts.OrganizationName, "id", token.ID())
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.OrganizationName, "id", token.ID())
	return token, nil
}

func (a *Application) ListAgentTokens(ctx context.Context, organizationName string) ([]*otf.AgentToken, error) {
	tokens, err := a.db.ListAgentTokens(ctx, organizationName)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organizationName)
		return nil, err
	}
	a.V(2).Info("listed agent tokens", "organization", organizationName)
	return tokens, nil
}

func (a *Application) GetAgentToken(ctx context.Context, id string) (*otf.AgentToken, error) {
	token, err := a.db.GetAgentToken(ctx, id)
	if err != nil {
		a.Error(err, "retrieving agent token", "id", id)
		return nil, err
	}
	a.V(0).Info("retrieved agent token", "organization", token.OrganizationName(), "id", token.ID())
	return token, nil
}

func (a *Application) DeleteAgentToken(ctx context.Context, id string) error {
	if err := a.db.DeleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "id", id)
		return err
	}
	a.V(0).Info("deleted agent token", "id", id)
	return nil
}
