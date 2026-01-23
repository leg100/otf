// Package client provides remote access to the OTF API
package client

import (
	"github.com/leg100/otf/internal/configversion"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

type (
	Client struct {
		*otfhttp.Client

		Organizations *organization.Client
		Workspaces    *workspace.Client
		Runs          *run.Client
		Teams         *team.Client
		Users         *user.Client
		States        *state.Client
		Configs       *configversion.Client
		Variables     *variable.Client
		Runners       *runner.Client
	}
)

func New(config otfhttp.ClientConfig) (*Client, error) {
	httpClient, err := otfhttp.NewClient(config)
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
		Client:        httpClient,
		Organizations: &organization.Client{Client: httpClient},
		Workspaces:    &workspace.Client{Client: httpClient},
		Runs:          &run.Client{Client: httpClient},
		Teams:         &team.Client{Client: httpClient},
		Users:         &user.Client{Client: httpClient},
		States:        &state.Client{Client: httpClient},
		Configs:       &configversion.Client{Client: httpClient},
		Variables:     &variable.Client{Client: httpClient},
		Runners:       &runner.Client{Client: httpClient},
	}
}
