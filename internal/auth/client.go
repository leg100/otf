package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
)

type (
	Client struct {
		internal.JSONAPIClient
	}
	teamMember struct {
		Username string `jsonapi:"primary,users"`
	}
)

// CreateUser creates a user via HTTP/JSONAPI. Options are ignored.
func (c *Client) CreateUser(ctx context.Context, username string, _ ...NewUserOption) (*User, error) {
	req, err := c.NewRequest("POST", "admin/users", &types.CreateUserOptions{
		Username: internal.String(username),
	})
	if err != nil {
		return nil, err
	}
	user := &types.User{}
	err = c.Do(ctx, req, user)
	if err != nil {
		return nil, err
	}
	return &User{ID: user.ID, Username: user.Username}, nil
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
func (c *Client) AddTeamMembership(ctx context.Context, opts TeamMembershipOptions) error {
	var members []*teamMember
	for _, name := range opts.Usernames {
		members = append(members, &teamMember{Username: name})
	}

	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(opts.TeamID))
	req, err := c.NewRequest("POST", u, members)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// RemoveTeamMembership removes users from a team via HTTP.
func (c *Client) RemoveTeamMembership(ctx context.Context, opts TeamMembershipOptions) error {
	var members []*teamMember
	for _, name := range opts.Usernames {
		members = append(members, &teamMember{Username: name})
	}

	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(opts.TeamID))
	req, err := c.NewRequest("DELETE", u, members)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// CreateTeam creates a team via HTTP/JSONAPI.
func (c *Client) CreateTeam(ctx context.Context, opts CreateTeamOptions) (*Team, error) {
	u := fmt.Sprintf("organizations/%s/teams", url.QueryEscape(opts.Organization))
	req, err := c.NewRequest("POST", u, &types.TeamCreateOptions{
		Name: internal.String(opts.Name),
	})
	if err != nil {
		return nil, err
	}
	team := &types.Team{}
	err = c.Do(ctx, req, team)
	if err != nil {
		return nil, err
	}
	return &Team{ID: team.ID, Name: team.Name}, nil
}

// GetTeam retrieves a team via HTTP/JSONAPI.
func (c *Client) GetTeam(ctx context.Context, organization, name string) (*Team, error) {
	u := fmt.Sprintf("organizations/%s/teams/%s", url.QueryEscape(organization), url.QueryEscape(name))
	req, err := c.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	team := &types.Team{}
	err = c.Do(ctx, req, team)
	if err != nil {
		return nil, err
	}
	return &Team{ID: team.ID, Name: team.Name}, nil
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
