package ui

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	h := &Handlers{AuthenticatorService: svc}
	h.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code, "output: %s", body) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}

func TestAdminLoginHandler(t *testing.T) {
	h := &Handlers{
		SiteToken: "secrettoken",
		Tokens:    &fakeTokensService{},
	}

	tests := []struct {
		name         string
		token        string
		wantRedirect string
	}{
		{
			name:         "valid token",
			token:        "secrettoken",
			wantRedirect: "/app/profile",
		},
		{
			name:         "invalid token",
			token:        "badtoken",
			wantRedirect: "/admin/login",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := strings.NewReader(url.Values{
				"token": {tt.token},
			}.Encode())

			r := httptest.NewRequest("POST", "/admin/login", form)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Add("Referer", "http://otf.server/admin/login")
			w := httptest.NewRecorder()
			h.adminLoginHandler(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, tt.wantRedirect, redirect.Path)
			}
		})
	}
}

type fakeLoginService struct {
	clients []*authenticator.OAuthClient
}

func (f *fakeLoginService) Clients() []*authenticator.OAuthClient {
	return f.clients
}
