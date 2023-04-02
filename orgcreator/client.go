package orgcreator

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/organization"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options OrganizationCreateOptions) (*organization.Organization, error) {
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
	org := jsonapi.Organization{}
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
