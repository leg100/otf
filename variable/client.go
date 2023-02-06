package variable

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

type Client struct {
	otf.JSONAPIClient
}

func (c *Client) ListVariables(ctx context.Context, workspaceID string) ([]otf.Variable, error) {
	u := fmt.Sprintf("workspaces/%s/vars", workspaceID)
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	list := &jsonapiList{}
	err = c.Do(ctx, req, list)
	if err != nil {
		return nil, err
	}

	var variables []otf.Variable
	for _, v := range list.Items {
		variables = append(variables, v.toVariable())
	}
	return variables, nil
}
