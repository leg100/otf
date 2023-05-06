package orgcreator

import (
	"context"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/organization"
)

type Client struct {
	internal.JSONAPIClient
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options OrganizationCreateOptions) (*organization.Organization, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	req, err := c.NewRequest("POST", "organizations", &types.OrganizationCreateOptions{
		Name:            options.Name,
		SessionRemember: options.SessionRemember,
		SessionTimeout:  options.SessionTimeout,
	})
	if err != nil {
		return nil, err
	}
	org := types.Organization{}
	err = c.Do(ctx, req, &org)
	if err != nil {
		return nil, err
	}
	return &organization.Organization{
		ID:              org.ExternalID,
		CreatedAt:       org.CreatedAt,
		Name:            org.Name,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}, nil
}
