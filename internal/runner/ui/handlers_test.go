package ui

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/stretchr/testify/assert"
)

func TestRunnerHandlers_createAgentPool(t *testing.T) {
	organization := organization.NewTestName(t)
	id := testutils.ParseID(t, "pool-123")
	svc := &fakeClient{
		pool: &runner.Pool{ID: id},
	}
	h := &Handlers{Client: svc}
	q := "/?organization_name=" + organization.String() + "&name=my-pool"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentPool(w, r)

	want := runner.CreateAgentPoolOptions{
		Name:         "my-pool",
		Organization: organization,
	}
	assert.Equal(t, want, svc.createAgentPoolOptions)
	testutils.AssertRedirect(t, w, paths.AgentPool(id))
}

func TestRunnerHandlers_listAgentPools(t *testing.T) {
	h := &Handlers{
		Client: &fakeClient{
			pool: &runner.Pool{ID: testutils.ParseID(t, "pool-123")},
		},
	}
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.listAgentPools(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestRunnerHandlers_createAgentToken(t *testing.T) {
	id := testutils.ParseID(t, "pool-123")
	h := &Handlers{
		Client: &fakeClient{},
	}
	q := "/?pool_id=pool-123&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentToken(w, r)

	testutils.AssertRedirect(t, w, paths.AgentPool(id))
}

func TestRunnerHandlers_deleteAgentToken(t *testing.T) {
	agentPoolID := resource.NewTfeID(resource.AgentPoolKind)

	h := &Handlers{
		Client: &fakeClient{
			at: &runner.AgentToken{
				AgentPoolID: agentPoolID,
			},
		},
	}
	q := fmt.Sprintf("/?token_id=%s", agentPoolID)
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()

	h.deleteAgentToken(w, r)

	testutils.AssertRedirect(t, w, paths.AgentPool(agentPoolID))
}

type fakeClient struct {
	Client
	pool                   *runner.Pool
	createAgentPoolOptions runner.CreateAgentPoolOptions
	at                     *runner.AgentToken
	token                  []byte
}

func (f *fakeClient) CreateAgentPool(ctx context.Context, opts runner.CreateAgentPoolOptions) (*runner.Pool, error) {
	f.createAgentPoolOptions = opts
	return f.pool, nil
}

func (f *fakeClient) UpdateAgentPool(ctx context.Context, poolID resource.TfeID, opts runner.UpdatePoolOptions) (*runner.Pool, error) {
	return nil, nil
}

func (f *fakeClient) ListAgentPoolsByOrganization(context.Context, organization.Name, runner.ListPoolOptions) ([]*runner.Pool, error) {
	return []*runner.Pool{f.pool}, nil
}

func (f *fakeClient) GetAgentPool(context.Context, resource.TfeID) (*runner.Pool, error) {
	return f.pool, nil
}

func (f *fakeClient) DeleteAgentPool(ctx context.Context, poolID resource.TfeID) (*runner.Pool, error) {
	return nil, nil
}

func (f *fakeClient) CreateAgentToken(context.Context, resource.TfeID, runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error) {
	return f.at, f.token, nil
}

func (f *fakeClient) ListAgentTokens(context.Context, resource.TfeID) ([]*runner.AgentToken, error) {
	return []*runner.AgentToken{f.at}, nil
}

func (f *fakeClient) GetAgentToken(context.Context, resource.TfeID) (*runner.AgentToken, error) {
	return f.at, nil
}

func (f *fakeClient) DeleteAgentToken(context.Context, resource.TfeID) (*runner.AgentToken, error) {
	return f.at, nil
}

func (f *fakeClient) ListRunners(ctx context.Context, opts runner.ListOptions) ([]*runner.RunnerMeta, error) {
	return nil, nil
}

func (f *fakeClient) DeleteRunner(ctx context.Context, runnerID resource.TfeID) error {
	return nil
}

func (f *fakeClient) Register(ctx context.Context, opts runner.RegisterRunnerOptions) (*runner.RunnerMeta, error) {
	return nil, nil
}
