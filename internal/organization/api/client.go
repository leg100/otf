package api

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

// Alias client to permit embedding it with other clients in a struct
// without a name clash.
type OrganizationClient = Client

type Client struct {
	*otfhttp.Client
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options organization.CreateOptions) (*organization.Organization, error) {
	if err := resource.ValidateName(options.Name); err != nil {
		return nil, err
	}
	req, err := c.NewRequest("POST", "organizations", &organization.CreateOptions{Name: options.Name})
	if err != nil {
		return nil, err
	}
	var org organization.Organization
	if err := c.Do(ctx, req, &org); err != nil {
		return nil, err
	}
	return &org, nil
}

// DeleteOrganization deletes an organization via http.
func (c *Client) DeleteOrganization(ctx context.Context, organization organization.Name) error {
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
