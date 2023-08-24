package variable

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type Client struct {
	internal.JSONAPIClient
}

func (c *Client) ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error) {
	u := fmt.Sprintf("workspaces/%s/vars", workspaceID)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	list := &types.VariableList{}
	err = c.Do(ctx, req, list)
	if err != nil {
		return nil, err
	}

	var variables []*Variable
	for _, v := range list.Items {
		variables = append(variables, &Variable{
			ID:          v.ID,
			Key:         v.Key,
			Value:       v.Value,
			Description: v.Description,
			Category:    VariableCategory(v.Category),
			Sensitive:   v.Sensitive,
			HCL:         v.HCL,
		})
	}
	return variables, nil
}
