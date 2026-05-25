// Package client provides remote access to the OTF API
package client

import (
	configversionapi "github.com/leg100/otf/internal/configversion/api"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	organizationapi "github.com/leg100/otf/internal/organization/api"
	runapi "github.com/leg100/otf/internal/run/api"
	runnerapi "github.com/leg100/otf/internal/runner/api"
	sshkeyapi "github.com/leg100/otf/internal/sshkey/api"
	stateapi "github.com/leg100/otf/internal/state/api"
	teamapi "github.com/leg100/otf/internal/team/api"
	userapi "github.com/leg100/otf/internal/user/api"
	variableapi "github.com/leg100/otf/internal/variable/api"
	workspaceapi "github.com/leg100/otf/internal/workspace/api"
)

type (
	Client struct {
		*otfhttp.Client

		*organizationapi.OrganizationClient
		*workspaceapi.WorkspaceClient
		*runapi.RunClient
		*teamapi.TeamClient
		*userapi.UserClient
		*stateapi.StateClient
		*configversionapi.ConfigClient
		*variableapi.VariableClient
		*runnerapi.RunnerClient
		*sshkeyapi.SSHKeyClient
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
		OrganizationClient: &organizationapi.Client{Client: httpClient},
		WorkspaceClient:    &workspaceapi.Client{Client: httpClient},
		RunClient:          &runapi.Client{Client: httpClient},
		TeamClient:         &teamapi.Client{Client: httpClient},
		UserClient:         &userapi.Client{Client: httpClient},
		StateClient:        &stateapi.Client{Client: httpClient},
		ConfigClient:       &configversionapi.Client{Client: httpClient},
		VariableClient:     &variableapi.Client{Client: httpClient},
		RunnerClient:       &runnerapi.Client{Client: httpClient},
		SSHKeyClient:       &sshkeyapi.Client{Client: httpClient},
	}
}
