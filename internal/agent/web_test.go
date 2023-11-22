package agent

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWebHandlers_createAgentPool(t *testing.T) {
	svc := &fakeService{
		pool: &Pool{ID: "pool-123"},
	}
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc:      svc,
	}
	q := "/?organization_name=acme-org&name=my-pool"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentPool(w, r)

	want := CreateAgentPoolOptions{
		Name:         "my-pool",
		Organization: "acme-org",
	}
	assert.Equal(t, want, svc.createAgentPoolOptions)
	testutils.AssertRedirect(t, w, paths.AgentPool("pool-123"))
}

func TestWebHandlers_listAgentPools(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc: &fakeService{
			pool: &Pool{ID: "pool-123"},
		},
	}
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.listAgentPools(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_createAgentToken(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc:      &fakeService{},
	}
	q := "/?organization_name=acme-org&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentToken(w, r)

	testutils.AssertRedirect(t, w, paths.AgentTokens("acme-org"))
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc: &fakeService{
			at: &agentToken{
				AgentPoolID: "pool-123",
			},
		},
	}
	q := "/?agent_token_id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()

	h.deleteAgentToken(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("pool-123"), redirect.Path)
	}
}
