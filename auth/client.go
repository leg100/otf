package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateUser creates a user via HTTP/JSONAPI. Options are ignored.
func (c *Client) CreateUser(ctx context.Context, username string, _ ...NewUserOption) (*User, error) {
	req, err := c.NewRequest("POST", "admin/users", &jsonapi.CreateUserOptions{
		Username: otf.String(username),
	})
	if err != nil {
		return nil, err
	}
	user := &jsonapi.User{}
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

// AddTeamMembership adds a user to a team via HTTP.
func (c *Client) AddTeamMembership(ctx context.Context, opts TeamMembershipOptions) error {
	u := fmt.Sprintf("teams/%s/memberships/%s", url.QueryEscape(opts.TeamID), url.QueryEscape(opts.Username))
	req, err := c.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}
	if err := c.Do(ctx, req, nil); err != nil {
		return err
	}
	return nil
}

// RemoveTeamMembership removes a user from a team via HTTP.
func (c *Client) RemoveTeamMembership(ctx context.Context, opts TeamMembershipOptions) error {
	u := fmt.Sprintf("teams/%s/memberships/%s", url.QueryEscape(opts.TeamID), url.QueryEscape(opts.Username))
	req, err := c.NewRequest("DELETE", u, nil)
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
	req, err := c.NewRequest("POST", u, &jsonapi.CreateTeamOptions{
		Name: otf.String(opts.Name),
	})
	if err != nil {
		return nil, err
	}
	team := &jsonapi.Team{}
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
	team := &jsonapi.Team{}
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

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *Client) CreateRegistrySession(ctx context.Context, opts CreateRegistrySessionOptions) (*RegistrySession, error) {
	req, err := c.NewRequest("POST", "registry/sessions/create", &jsonapi.RegistrySessionCreateOptions{
		Organization: opts.Organization,
	})
	if err != nil {
		return nil, err
	}
	session := &jsonapi.RegistrySession{}
	err = c.Do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return &RegistrySession{
		Organization: session.OrganizationName,
		Token:        session.Token,
	}, nil
}

func (c *Client) CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) (*AgentToken, error) {
	req, err := c.NewRequest("POST", "agent/create", &jsonapi.AgentTokenCreateOptions{
		Description:  options.Description,
		Organization: options.Organization,
	})
	if err != nil {
		return nil, err
	}
	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}
	return &AgentToken{ID: at.ID, Token: *at.Token, Organization: at.Organization}, nil
}

func (c *Client) GetAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	req, err := c.NewRequest("GET", "agent/details", nil)
	if err != nil {
		return nil, err
	}

	at := &jsonapi.AgentToken{}
	err = c.Do(ctx, req, at)
	if err != nil {
		return nil, err
	}

	return &AgentToken{ID: at.ID, Organization: at.Organization}, nil
}
