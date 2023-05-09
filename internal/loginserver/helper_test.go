package loginserver

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/tokens"
	"github.com/stretchr/testify/require"
)

func fakeServer(t *testing.T, secret []byte) *server {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	srv, err := NewServer(Options{
		Secret:        secret,
		Renderer:      renderer,
		TokensService: &fakeTokenService{},
	})
	require.NoError(t, err)
	return srv
}

type fakeTokenService struct {
	tokens.TokensService
}

func (a *fakeTokenService) CreateUserToken(ctx context.Context, opts tokens.CreateUserTokenOptions) (*tokens.UserToken, []byte, error) {
	return nil, nil, nil
}
