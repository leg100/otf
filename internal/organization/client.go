package organization

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfhttp.Client
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options CreateOptions) (*Organization, error) {
	if err := resource.ValidateName(options.Name); err != nil {
		return nil, err
	}
	req, err := c.NewRequest("POST", "organizations", &CreateOptions{Name: options.Name})
	if err != nil {
		return nil, err
	}
	var org Organization
	if err := c.Do(ctx, req, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// DeleteOrganization deletes an organization via http.
func (c *Client) DeleteOrganization(ctx context.Context, organization Name) error {
	u := fmt.Sprintf("organizations/%s", organization)
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
