package organization

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	internal.JSONAPIClient
}

// CreateOrganization creates a new organization with the given options.
func (c *Client) CreateOrganization(ctx context.Context, options CreateOptions) (*Organization, error) {
	if err := resource.ValidateName(options.Name); err != nil {
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
	return &Organization{
		ID:              org.ExternalID,
		CreatedAt:       org.CreatedAt,
		Name:            org.Name,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}, nil
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
