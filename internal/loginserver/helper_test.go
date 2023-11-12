package loginserver

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/require"
)

func fakeServer(t *testing.T, secret []byte) *server {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	srv, err := NewServer(Options{
		Secret:      secret,
		Renderer:    renderer,
		UserService: &fakeUserService{},
	})
	require.NoError(t, err)
	return srv
}

type fakeUserService struct {
	user.UserService
}

func (a *fakeUserService) CreateUserToken(context.Context, user.CreateUserTokenOptions) (*user.UserToken, []byte, error) {
	return nil, nil, nil
}
