package loginserver

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
)

func fakeServer(t *testing.T, secret []byte) *server {
	return &server{
		secret:   secret,
		Renderer: testutils.NewRenderer(t),
		users:    &fakeUserService{},
	}
}

type fakeUserService struct{}

func (a *fakeUserService) CreateToken(context.Context, user.CreateUserTokenOptions) (*user.UserToken, []byte, error) {
	return nil, nil, nil
}
