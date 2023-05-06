package organization

import (
	"context"
	"fmt"

	internal "github.com/leg100/otf"
)

type Client struct {
	internal.JSONAPIClient
}

// DeleteOrganization deletes an organization via http.
func (c *Client) DeleteOrganization(ctx context.Context, organization string) error {
	u := fmt.Sprintf("organizations/%s", organization)
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	err = c.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}
