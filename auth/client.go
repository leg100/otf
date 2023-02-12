package auth

import (
	"context"
	"path"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *Client) CreateRegistrySession(ctx context.Context, organization string) (string, error) {
	path := path.Join("organizations", organization, "registry/sessions/create")
	req, err := c.NewRequest("POST", path, &jsonapi.RegistrySessionCreateOptions{
		OrganizationName: organization,
	})
	if err != nil {
		return "", err
	}
	session := &jsonapi.RegistrySession{}
	err = c.Do(ctx, req, session)
	if err != nil {
		return "", err
	}
	return session.Token, nil
}

func (c *Client) CreateAgentToken(ctx context.Context, options otf.CreateAgentTokenOptions) (otf.AgentToken, error) {
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
	return &agentToken{id: at.ID, token: *at.Token, organization: at.Organization}, nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*agentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return &agentToken{id: at.ID, organization: at.Organization}, nil
}
