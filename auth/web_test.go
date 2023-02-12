package auth

import (
	"context"
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
	app *fakeApp
	otf.Renderer
}

func newFakeWeb(t *testing.T, app *fakeApp) *web {
	renderer, err := html.NewViewEngine(true)
	require.NoError(t, err)
	return &web{
		app:      app,
		Renderer: renderer,
	}
}

func TestSessionHandlers(t *testing.T) {
	user := otf.NewTestUser(t)
	active := otf.NewTestSession(t, user.ID())
	other := otf.NewTestSession(t, user.ID())

	app := newFakeWebApp(t, &fakeSessionHandlerApp{sessions: []*otf.Session{active, other}})

	t.Run("list sessions", func(t *testing.T) {
		// add user and active session to request
		r := httptest.NewRequest("GET", "/sessions", nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		r = r.WithContext(addSessionCtx(r.Context(), active))

		w := httptest.NewRecorder()
		app.sessionsHandler(w, r)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("revoke session", func(t *testing.T) {
		form := strings.NewReader(url.Values{
			"token": {"asklfdkljfj"},
		}.Encode())

		r := httptest.NewRequest("POST", "/sessions/delete", form)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		app.revokeSessionHandler(w, r)

		assert.Equal(t, 302, w.Code)
	})
}

func TestAdminLoginHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeAdminLoginHandlerApp{}, withFakeSiteToken("secrettoken"))

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

type fakeSessionHandlerApp struct {
	sessions []*otf.Session
	otf.Application
}

func (f *fakeSessionHandlerApp) ListSessions(context.Context, string) ([]*otf.Session, error) {
	return f.sessions, nil
}

func (f *fakeSessionHandlerApp) DeleteSession(context.Context, string) error {
	return nil
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	app := newFakeWebApp(t, nil, withAuthenticators([]*authenticator{
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
	}))

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	app.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}
