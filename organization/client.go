package organization

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options OrganizationCreateOptions) (*Organization, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	req, err := c.NewRequest("POST", "organizations", &jsonapi.OrganizationCreateOptions{
		Name:            options.Name,
		SessionRemember: options.SessionRemember,
		SessionTimeout:  options.SessionTimeout,
	})
	if err != nil {
		return nil, err
	}
	org := &JSONAPIOrganization{}
	err = c.Do(ctx, req, org)
	if err != nil {
		return nil, err
	}
	return org.toOrganization(), nil
}
