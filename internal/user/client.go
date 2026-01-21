package user

import (
	"context"
	"fmt"
	"net/url"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

type Client struct {
	*otfhttp.Client
}

// Create creates a user via HTTP/JSONAPI. Options are ignored.
func (c *Client) Create(ctx context.Context, username string, _ ...NewUserOption) (*User, error) {
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

// Delete deletes a user via HTTP/JSONAPI.
func (c *Client) Delete(ctx context.Context, username Username) error {
	u := fmt.Sprintf("admin/users/%s", url.QueryEscape(username.String()))
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
func (c *Client) AddTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []Username) error {
	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(teamID.String()))
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
func (c *Client) RemoveTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []Username) error {
	u := fmt.Sprintf("teams/%s/relationships/users", url.QueryEscape(teamID.String()))
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
