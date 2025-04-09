package variable

import (
	"context"
	"fmt"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfapi.Client
}

func (c *Client) ListEffectiveVariables(ctx context.Context, runID resource.TfeID) ([]*Variable, error) {
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
