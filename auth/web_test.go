package auth

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWeb struct {
	app *fakeService
	otf.Renderer
}

func newFakeWeb(t *testing.T, svc service) *web {
	renderer, err := html.NewViewEngine(true)
	require.NoError(t, err)
	return &web{
		svc:      svc,
		Renderer: renderer,
	}
}

func TestAdminLoginHandler(t *testing.T) {
	app := newFakeWeb(t, &fakeService{
		sessionService: &fakeSessionService{},
	})
	app.siteToken = "secrettoken"

	tests := []struct {
		name         string
		token        string
		wantRedirect string
	}{
		{
			name:         "valid token",
			token:        "secrettoken",
			wantRedirect: "/profile",
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
			w := httptest.NewRecorder()
			app.adminLoginHandler(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, tt.wantRedirect, redirect.Path)
			}
		})
	}
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	app := newFakeWeb(t, &fakeService{})
	app.authenticators = []*authenticator{
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud1"},
			},
		},
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud2"},
			},
		},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	app.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}
