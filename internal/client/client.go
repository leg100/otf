// Package client provides remote access to the OTF API
package client

import (
	"github.com/leg100/otf/internal/configversion"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sshkey"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type (
	Client struct {
		*otfhttp.Client

		*organization.OrganizationClient
		*workspace.WorkspaceClient
		*run.RunClient
		*team.TeamClient
		*user.UserClient
		*state.StateClient
		*configversion.ConfigClient
		*variable.VariableClient
		*runner.RunnerClient
		*sshkey.SSHKeyClient
	}
)

func New(
	logger logr.Logger,
	url string,
	token string,
) (*Client, error) {
	httpClient, err := otfhttp.NewClient(otfhttp.ClientConfig{
		URL:           url,
		Logger:        logger,
		Token:         token,
		RetryRequests: true,
	})
	if err != nil {
		return nil, err
	}
	return (&Client{}).new(httpClient), nil
}

// UseToken returns a shallow copy of the client with a different auth token.
func (c *Client) UseToken(token string) *Client {
	c2 := *c.Client
	c2.Token = token
	return c.new(&c2)
}

func (c *Client) new(httpClient *otfhttp.Client) *Client {
	return &Client{
		Client:             httpClient,
		OrganizationClient: &organization.Client{Client: httpClient},
		WorkspaceClient:    &workspace.Client{Client: httpClient},
		RunClient:          &run.Client{Client: httpClient},
		TeamClient:         &team.Client{Client: httpClient},
		UserClient:         &user.Client{Client: httpClient},
		StateClient:        &state.Client{Client: httpClient},
		ConfigClient:       &configversion.Client{Client: httpClient},
		VariableClient:     &variable.Client{Client: httpClient},
		RunnerClient:       &runner.Client{Client: httpClient},
		SSHKeyClient:       &sshkey.Client{Client: httpClient},
	}
}
