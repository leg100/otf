package agenttoken

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

func TestAgentToken_NewHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		org: org,
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newAgentToken(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_CreateHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	token := otf.NewTestAgentToken(t, org)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		token: token,
	})

	q := "/?organization_name=acme&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.createAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme"), redirect.Path)
	}
}

func TestAgentToken_ListHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	token := otf.NewTestAgentToken(t, org)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		org:   org,
		token: token,
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listAgentTokens(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		token: otf.NewTestAgentToken(t, org),
	})

	q := "/?agent_token_id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	app.deleteAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens(org.Name()), redirect.Path)
	}
}

type fakeAgentTokenHandlerApp struct {
	org   *otf.Organization
	token *otf.AgentToken

	otf.Application
}

func (f *fakeAgentTokenHandlerApp) GetOrganization(context.Context, string) (*otf.Organization, error) {
	return f.org, nil
}

func (f *fakeAgentTokenHandlerApp) CreateAgentToken(context.Context, otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return f.token, nil
}

func (f *fakeAgentTokenHandlerApp) ListAgentTokens(context.Context, string) ([]*otf.AgentToken, error) {
	return []*otf.AgentToken{f.token}, nil
}

func (f *fakeAgentTokenHandlerApp) DeleteAgentToken(context.Context, string) (*otf.AgentToken, error) {
	return f.token, nil
}
