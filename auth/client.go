package auth

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *Client) CreateRegistrySession(ctx context.Context, opts CreateRegistrySessionOptions) (*RegistrySession, error) {
	req, err := c.NewRequest("POST", "registry/sessions/create", &jsonapi.RegistrySessionCreateOptions{
		Organization: opts.Organization,
	})
	if err != nil {
		return nil, err
	}
	session := &jsonapi.RegistrySession{}
	err = c.Do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return &RegistrySession{
		Organization: session.OrganizationName,
		Token:        session.Token,
	}, nil
}

func (c *Client) CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) (*AgentToken, error) {
	req, err := c.NewRequest("POST", "agent/create", &jsonapi.AgentTokenCreateOptions{
		Description:  options.Description,
		Organization: options.Organization,
	})
	if err != nil {
		return nil, err
	}
	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}
	return &AgentToken{ID: at.ID, Token: *at.Token, Organization: at.Organization}, nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return &AgentToken{ID: at.ID, Organization: at.Organization}, nil
}
