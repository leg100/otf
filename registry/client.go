package registry

import (
	"context"
	"path"

	"github.com/leg100/otf"
)

type Client struct {
	otf.JSONAPIClient
}

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *Client) CreateRegistrySession(ctx context.Context, organization string) (otf.RegistrySession, error) {
	path := path.Join("organizations", organization, "registry/sessions/create")
	req, err := c.NewRequest("POST", path, &jsonapiCreateOptions{
		OrganizationName: organization,
	})
	if err != nil {
		return nil, err
	}
	session := &jsonapiSession{}
	err = c.Do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return session.toSession(), nil
}
