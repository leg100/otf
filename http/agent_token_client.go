package http

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (c *client) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	req, err := c.newRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &dto.AgentToken{}
	err = c.do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalAgentTokenJSONAPI(at), nil
}
