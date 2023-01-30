package http

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// CreateOrganization creates a new organization with the given options.
func (c *client) CreateOrganization(ctx context.Context, options otf.OrganizationCreateOptions) (*otf.Organization, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "organizations", &jsonapi.OrganizationCreateOptions{
		Name:            options.Name,
		SessionRemember: options.SessionRemember,
		SessionTimeout:  options.SessionTimeout,
	})
	if err != nil {
		return nil, err
	}
	org := &jsonapi.Organization{}
	err = c.do(ctx, req, org)
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalOrganizationJSONAPI(org), nil
}
