package auth

import (
	"context"
	"fmt"
	"net/url"

	otfapi "github.com/leg100/otf/internal/api"
)

type (
	Client struct {
		*otfapi.Client

		AuthService
	}
)

// CreateUser creates a user via HTTP/JSONAPI. Options are ignored.
func (c *Client) CreateUser(ctx context.Context, username string, _ ...NewUserOption) (*User, error) {
	req, err := c.NewRequest("POST", "admin/users", &CreateUserOptions{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	var user User
	if err := c.Do(ctx, req, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser deletes a user via HTTP/JSONAPI.
func (c *Client) DeleteUser(ctx context.Context, username string) error {
	u := fmt.Sprintf("admin/users/%s", url.QueryEscape(username))
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// AddTeamMembership adds users to a team via HTTP.
func (c *Client) AddTeamMembership(ctx context.Context, teamID string, usernames []string) error {
	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(teamID))
	req, err := c.NewRequest("POST", u, &modifyTeamMembershipOptions{
		Usernames: usernames,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// RemoveTeamMembership removes users from a team via HTTP.
func (c *Client) RemoveTeamMembership(ctx context.Context, teamID string, usernames []string) error {
	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(teamID))
	req, err := c.NewRequest("DELETE", u, &modifyTeamMembershipOptions{
		Usernames: usernames,
	})
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// CreateTeam creates a team via HTTP/JSONAPI.
func (c *Client) CreateTeam(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error) {
	// validate params
	if _, err := newTeam(organization, opts); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("organizations/%s/teams", url.QueryEscape(organization))
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

// GetTeam retrieves a team via HTTP/JSONAPI.
func (c *Client) GetTeam(ctx context.Context, organization, name string) (*Team, error) {
	u := fmt.Sprintf("organizations/%s/teams/%s", url.QueryEscape(organization), url.QueryEscape(name))
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

// DeleteTeam deletes a team via HTTP/JSONAPI.
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	u := fmt.Sprintf("teams/%s", url.QueryEscape(id))
	req, err := c.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}
