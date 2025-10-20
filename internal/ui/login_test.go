package ui

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/github"
	"github.com/stretchr/testify/assert"
)

type fakeLoginService struct {
	clients []*authenticator.OAuthClient
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	svc := &fakeLoginService{}

	svc.clients = []*authenticator.OAuthClient{
		{OAuthConfig: authenticator.OAuthConfig{Name: "cloud1", Icon: github.Icon()}},
		{OAuthConfig: authenticator.OAuthConfig{Name: "cloud2", Icon: github.Icon()}},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	h := &loginHandler{loginService: svc}
	h.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code, "output: %s", body) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}

func (f *fakeLoginService) Clients() []*authenticator.OAuthClient {
	return f.clients
}
