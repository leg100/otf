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

func (svc *fakeUserService) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	if spec.AuthenticationToken == nil {
		return nil, otf.ErrResourceNotFound
	}
	if *spec.AuthenticationToken != svc.token {
		return nil, otf.ErrResourceNotFound
	}
	return nil, nil
}
