package loginserver

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html"
	"github.com/stretchr/testify/require"
)

func fakeServer(t *testing.T, secret []byte) *server {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	srv, err := NewServer(Options{
		Secret:      secret,
		Renderer:    renderer,
		AuthService: &fakeTokenService{},
	})
	require.NoError(t, err)
	return srv
}

type fakeTokenService struct {
	auth.AuthService
}

func (a *fakeTokenService) CreateUserToken(ctx context.Context, opts auth.CreateUserTokenOptions) (*auth.UserToken, []byte, error) {
	return nil, nil, nil
}
