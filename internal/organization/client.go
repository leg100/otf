package organization

import (
	"context"

	"github.com/leg100/otf/internal/apigen"
	otfhttp "github.com/leg100/otf/internal/http"
)

type Client struct {
	*otfhttp.Client
}

// DeleteOrganization deletes an organization via http.
func (c *Client) DeleteOrganization(ctx context.Context, organization string) error {
	return c.Client.DeleteOrganization(ctx, apigen.DeleteOrganizationParams{
		Name: organization,
	})
}
