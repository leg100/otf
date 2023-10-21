package variable

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
)

type Client struct {
	internal.JSONAPIClient
}

func (c *Client) ListEffectiveVariables(ctx context.Context, runID string) ([]*Variable, error) {
	u := fmt.Sprintf("vars/effective/%s", runID)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var list []*Variable
	if err := c.Do(ctx, req, &list); err != nil {
		return nil, err
	}

	return list, nil
}
