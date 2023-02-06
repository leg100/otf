package agenttoken

import (
	"context"

	"github.com/leg100/otf"
)

type Client struct {
	otf.JSONAPIClient
}

func (c *Client) CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) (*AgentToken, error) {
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

func (c *Client) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
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
