package tokens

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenWeb(t *testing.T) {
	user := &auth.User{Username: uuid.NewString()}

	t.Run("new", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()

		web.newUserToken(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("create", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.createUserToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})

	t.Run("list", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.userTokens(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("delete", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?id=token-123"
		r := httptest.NewRequest("POST", q, nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.deleteUserToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})
}

func TestAgentToken_NewHandler(t *testing.T) {
	web := newTestAgentTokenHandlers(t, "acme-org")
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	web.newAgentToken(w, r)

	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_CreateHandler(t *testing.T) {
	web := newTestAgentTokenHandlers(t, "acme-org")
	q := "/?organization_name=acme-org&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	web.createAgentToken(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
	}
}

func TestAgentToken_ListHandler(t *testing.T) {
	web := newTestAgentTokenHandlers(t, "acme-org")
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	web.listAgentTokens(w, r)

	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	web := newTestAgentTokenHandlers(t, "acme-org")
	q := "/?agent_token_id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()

	web.deleteAgentToken(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
	}
}

func newFakeWeb(t *testing.T, svc TokensService) *webHandlers {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &webHandlers{
		svc:      svc,
		Renderer: renderer,
	}
}

func TestAdminLoginHandler(t *testing.T) {
	app := newFakeWeb(t, &fakeService{})
	app.siteToken = "secrettoken"

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
			w := httptest.NewRecorder()
			app.adminLogin(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, tt.wantRedirect, redirect.Path)
			}
		})
	}
}

func newTestAgentTokenHandlers(t *testing.T, org string) *webHandlers {
	return newFakeWeb(t, &fakeService{
		agentToken: NewTestAgentToken(t, org),
	})
}

func newTestTokenHandlers(t *testing.T, org string) *webHandlers {
	return newFakeWeb(t, &fakeService{
		userToken: NewTestToken(t, org),
	})
}
