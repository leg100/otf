package http

import (
	"context"
	"path"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *client) CreateRegistrySession(ctx context.Context, organization string) (*otf.RegistrySession, error) {
	path := path.Join("organizations", organization, "registry/sessions/create")
	req, err := c.newRequest("POST", path, &jsonapi.RegistrySessionCreateOptions{
		OrganizationName: organization,
	})
	if err != nil {
		return nil, err
	}
	session := &jsonapi.RegistrySession{}
	err = c.do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalRegistrySessionJSONAPI(session), nil
}
