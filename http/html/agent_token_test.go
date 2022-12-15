package html

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
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
		tokens: []*otf.AgentToken{token},
	})

	q := "/?organization_name=acme&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.createAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, agentTokensPath("acme"), redirect.Path)
	}
}

func TestAgentToken_ListHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	token := otf.NewTestAgentToken(t, org)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		org:    org,
		tokens: []*otf.AgentToken{token},
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
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{})

	q := "/?organization_name=acme&id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	app.deleteAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, agentTokensPath("acme"), redirect.Path)
	}
}

type fakeAgentTokenHandlerApp struct {
	org    *otf.Organization
	tokens []*otf.AgentToken

	otf.Application
}

func (f *fakeAgentTokenHandlerApp) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	return f.org, nil
}

func (f *fakeAgentTokenHandlerApp) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return f.tokens[0], nil
}

func (f *fakeAgentTokenHandlerApp) ListAgentTokens(ctx context.Context, organization string) ([]*otf.AgentToken, error) {
	return f.tokens, nil
}

func (f *fakeAgentTokenHandlerApp) DeleteAgentToken(ctx context.Context, id, organization string) error {
	return nil
}
