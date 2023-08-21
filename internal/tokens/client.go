package tokens

import (
	"bytes"
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type Client struct {
	internal.JSONAPIClient

	// client doesn't implement all of service yet
	TokensService
}

func NewClient(httpClient *http.Client) (*Client, error) {
	return &Client{JSONAPIClient: httpClient}, nil
}

// CreateRunToken creates a run token via HTTP/JSONAPI
func (c *Client) CreateRunToken(ctx context.Context, opts CreateRunTokenOptions) ([]byte, error) {
	req, err := c.NewRequest("POST", "tokens/run/create", &types.CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) ([]byte, error) {
	req, err := c.NewRequest("POST", "agent/create", &types.AgentTokenCreateOptions{
		Description:  options.Description,
		Organization: options.Organization,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = c.Do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &types.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return &AgentToken{ID: at.ID, Organization: at.Organization}, nil
}
