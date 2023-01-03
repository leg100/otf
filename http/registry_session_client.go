package http

import (
	"context"
	"path"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

// CreateRegistrySession creates a registry session via HTTP/JSONAPI
func (c *client) CreateRegistrySession(ctx context.Context, organization string) (*otf.RegistrySession, error) {
	path := path.Join("organizations", organization, "registry/sessions/create")
	req, err := c.newRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}
	session := &dto.RegistrySession{}
	err = c.do(ctx, req, session)
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalRegistrySessionJSONAPI(session), nil
}
