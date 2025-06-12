package authenticator

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticatorService(t *testing.T) {
	opts := Options{
		Logger:          logr.Discard(),
		HostnameService: internal.NewHostnameService("fake-host.org"),
	}
	_, err := NewAuthenticatorService(context.Background(), opts)
	require.NoError(t, err)
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	svc := &Service{}

	svc.clients = []*OAuthClient{
		{OAuthConfig: OAuthConfig{Name: "cloud1", Icon: oidcIcon()}},
		{OAuthConfig: OAuthConfig{Name: "cloud2", Icon: oidcIcon()}},
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
