package loginserver

import (
	"testing"

	"github.com/leg100/otf/internal/http/html"
	"github.com/stretchr/testify/require"
)

func fakeServer(t *testing.T, secret []byte) *server {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	srv, err := NewServer(secret, renderer)
	require.NoError(t, err)
	return srv
}
