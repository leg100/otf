package agent

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAgentToken_NewHandler(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
	}
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.newAgentToken(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestAgentToken_CreateHandler(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc:      &fakeAgentTokenService{},
	}
	q := "/?organization_name=acme-org&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentToken(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
	}
}

func TestAgentToken_ListHandler(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc: &fakeAgentTokenService{
			agentToken: &AgentToken{},
		},
	}
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.listAgentTokens(w, r)

	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		svc: &fakeAgentTokenService{
			agentToken: &AgentToken{
				Organization: "acme-org",
			},
		},
	}
	q := "/?agent_token_id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()

	h.deleteAgentToken(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme-org"), redirect.Path)
	}
}

type fakeAgentTokenService struct {
	agentToken *AgentToken
	token      []byte

	AgentTokenService
}

func (f *fakeAgentTokenService) CreateAgentToken(context.Context, CreateAgentTokenOptions) ([]byte, error) {
	return f.token, nil
}

func (f *fakeAgentTokenService) ListAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.agentToken}, nil
}

func (f *fakeAgentTokenService) DeleteAgentToken(context.Context, string) (*AgentToken, error) {
	return f.agentToken, nil
}
