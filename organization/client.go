package organization

import (
	"context"
	"fmt"

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
	org := jsonapi.Organization{}
	err = c.Do(ctx, req, &org)
	if err != nil {
		return nil, err
	}
	return newFromJSONAPI(org), nil
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
