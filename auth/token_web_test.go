package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

func TestTokenWeb(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		web := newTestAgentTokenHandlers(t, "acme-org")
		q := "/?organization_name=acme-org"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()

		web.newAgentToken(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("create", func(t *testing.T) {
		web := newTestAgentTokenHandlers(t, "acme-org")
		q := "/?organization_name=acme-org&description=lorem-ipsum-etc"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()

		web.createAgentToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
		}
	})

	t.Run("list", func(t *testing.T) {
		web := newTestAgentTokenHandlers(t, "acme-org")
		q := "/?organization_name=acme-org"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()

		web.listAgentTokens(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("delete", func(t *testing.T) {
		web := newTestAgentTokenHandlers(t, "acme-org")
		q := "/?agent_token_id=at-123"
		r := httptest.NewRequest("POST", q, nil)
		w := httptest.NewRecorder()

		web.deleteAgentToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
		}
	})
}

func newTestTokenHandlers(t *testing.T, org string) *webHandlers {
	return newFakeWeb(t, &fakeService{
		tokenService: &fakeTokenService{
			token: NewTestToken(t, org),
		},
	})
}
