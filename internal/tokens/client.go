package tokens

import (
	"bytes"
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http"
)

type Client struct {
	internal.JSONAPIClient

	// client doesn't implement all of service yet
	TokensService
}

func NewClient(httpClient *http.Client) (*Client, error) {
	return &Client{JSONAPIClient: httpClient}, nil
}

func (c *Client) CreateRunToken(ctx context.Context, opts CreateRunTokenOptions) ([]byte, error) {
	req, err := c.NewRequest("POST", "tokens/run/create", opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	req, err := c.NewRequest("POST", "agent/create", opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}
	var at AgentToken
	if err := c.Do(ctx, req, &at); err != nil {
		return nil, err
	}
	return &at, nil
}
