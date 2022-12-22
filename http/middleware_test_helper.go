package http

import (
	"context"

	"github.com/leg100/otf"
)

type fakeUserService struct {
	// auth token to valid user
	token string
	otf.UserService
}

func (svc *fakeUserService) GetUser(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	if spec.AuthenticationToken == nil {
		return nil, otf.ErrResourceNotFound
	}
	if *spec.AuthenticationToken != svc.token {
		return nil, otf.ErrResourceNotFound
	}
	return nil, nil
}

type fakeAgentTokenService struct {
	token string
	otf.AgentTokenService
}

func (f *fakeAgentTokenService) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	if token != f.token {
		return nil, otf.ErrResourceNotFound
	}

	return otf.NewAgentToken(otf.CreateAgentTokenOptions{
		OrganizationName: "fake-org",
		Description:      "fake token",
	})
}
