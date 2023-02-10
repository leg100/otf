package auth

import (
	"context"
	"path"

	"github.com/leg100/otf"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *Client) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	path := path.Join("organizations", organization, "registry/sessions/create")
	req, err := c.NewRequest("POST", path, &jsonapiCreateOptions{
		OrganizationName: organization,
	})
	if err != nil {
		return nil, err
	}
	session := &jsonapiSession{}
	err = c.Do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return session.toSession(), nil
}
package agenttoken

import (
	"context"

	"github.com/leg100/otf"
)

type Client struct {
	otf.JSONAPIClient
}

func (c *Client) CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) (*agentToken, error) {
	req, err := c.NewRequest("POST", "agent/create", &jsonapiCreateOptions{
		Description:  options.Description,
		Organization: options.Organization,
	})
	if err != nil {
		return nil, err
	}
	at := &jsonapiAgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}
	return UnmarshalAgentTokenJSONAPI(at), nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*agentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &jsonapiAgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return UnmarshalAgentTokenJSONAPI(at), nil
}
