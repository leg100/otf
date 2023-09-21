package authenticator

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticatorService(t *testing.T) {
	opts := Options{
		Logger:          logr.Discard(),
		HostnameService: internal.NewHostnameService("fake-host.org"),
		OpaqueHandlerConfigs: []OpaqueHandlerConfig{
			{
				Kind: cloud.GithubKind,
				OAuthConfig: OAuthConfig{
					ClientID:     "id-1",
					ClientSecret: "secret-1",
				},
			},
			{
				OAuthConfig: OAuthConfig{
					ClientID:     "id-2",
					ClientSecret: "secret-2",
				},
			},
			{
				// should be skipped
				OAuthConfig: OAuthConfig{},
			},
		},
	}
	got, err := NewAuthenticatorService(context.Background(), opts)
	require.NoError(t, err)
	assert.Equal(t, 2, len(got.clients))
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	svc := &service{Renderer: renderer}

	svc.clients = []*OAuthClient{
		{OAuthConfig: OAuthConfig{Name: "cloud1"}},
		{OAuthConfig: OAuthConfig{Name: "cloud2"}},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	svc.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code, "output: %s", body) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}
