package variable

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
}

func (c *Client) ListVariables(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
	u := fmt.Sprintf("workspaces/%s/vars", workspaceID)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	list := &jsonapi.VariableList{}
	err = c.Do(ctx, req, list)
	if err != nil {
		return nil, err
	}

	var variables []*otf.Variable
	for _, v := range list.Items {
		variables = append(variables, &otf.Variable{
			ID:          v.ID,
			Key:         v.Key,
			Value:       v.Value,
			Description: v.Description,
			Category:    otf.VariableCategory(v.Category),
			Sensitive:   v.Sensitive,
			HCL:         v.HCL,
			WorkspaceID: v.Workspace.ID,
		})
	}
	return variables, nil
}
