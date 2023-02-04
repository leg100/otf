package http

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (c *Client) CreateAgentToken(ctx context.Context, options otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
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
	return otf.UnmarshalAgentTokenJSONAPI(at), nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalAgentTokenJSONAPI(at), nil
}
