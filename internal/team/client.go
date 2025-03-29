package team

import (
	"context"
	"fmt"
	"net/url"

	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfapi.Client
}

// Create creates a team via HTTP/JSONAPI.
func (c *Client) Create(ctx context.Context, organization resource.OrganizationName, opts CreateTeamOptions) (*Team, error) {
	// validate params
	if _, err := newTeam(organization, opts); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("organizations/%s/teams", url.QueryEscape(organization.String()))
	req, err := c.NewRequest("POST", u, &opts)
	if err != nil {
		return nil, err
	}
	var team Team
	if err := c.Do(ctx, req, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

// Get retrieves a team via HTTP/JSONAPI.
func (c *Client) Get(ctx context.Context, organization resource.OrganizationName, name string) (*Team, error) {
	u := fmt.Sprintf("organizations/%s/teams/%s", url.QueryEscape(organization.String()), url.QueryEscape(name))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var team Team
	if err := c.Do(ctx, req, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

// Delete deletes a team via HTTP/JSONAPI.
func (c *Client) Delete(ctx context.Context, id resource.ID) error {
	u := fmt.Sprintf("teams/%s", url.QueryEscape(id.String()))
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
