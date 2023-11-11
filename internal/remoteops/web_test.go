package remoteops

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func newFakeWeb(t *testing.T, svc agentTokenService) *webHandlers {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &webHandlers{
		svc:      svc,
		Renderer: renderer,
	}
}

func newTestAgentTokenHandlers(t *testing.T, org string) *webHandlers {
	return newFakeWeb(t, &fakeService{
		agentToken: NewTestAgentToken(t, org),
	})
}

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, _, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: CreateAgentTokenOptions{
			Organization: org,
			Description:  "lorem ipsum...",
		},
		key: newTestJWK(t, testutils.NewSecret(t)),
	})
	require.NoError(t, err)
	return token
}
