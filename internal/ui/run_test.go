package ui

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsHandlers(t *testing.T) {
	ws := workspace.NewTestWorkspace(t, nil)
	workspaces := &fakeWorkspaceClient{
		ws: ws,
	}
	cv := configversion.NewConfigurationVersion(ws.ID, configversion.CreateOptions{})
	run, err := run.NewRun(ws, cv, run.CreateOptions{})
	require.NoError(t, err)

	h := &Handlers{
		Workspaces: workspaces,
		Runs: &fakeRunClient{
			run: run,
		},
		Configs:    &fakeConfigsClient{},
		Authorizer: authz.NewAllowAllAuthorizer(),
		templates: &templates{
			workspaces: workspaces,
			configs:    &fakeConfigsClient{},
		},
	}
	user := &user.User{ID: resource.NewTfeID(resource.UserKind)}

	t.Run("list runs", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page=1", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.listRuns(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("get run", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?run_id=run-123", nil)
		w := httptest.NewRecorder()
		h.getRun(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})
}

func TestRuns_CancelHandler(t *testing.T) {
	run := &run.Run{ID: testutils.ParseID(t, "run-1")}
	h := &Handlers{
		Runs: &fakeRunClient{},
	}

	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
	w := httptest.NewRecorder()
	h.cancelRun(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Equal(t, paths.Run(run.ID), w.Header().Get("HX-Redirect"))
}

func TestRuns_ForceCancelHandler(t *testing.T) {
	run := &run.Run{ID: testutils.ParseID(t, "run-1")}
	h := &Handlers{
		Runs: &fakeRunClient{},
	}

	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
	w := httptest.NewRecorder()
	h.forceCancelRun(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Equal(t, paths.Run(run.ID), w.Header().Get("HX-Redirect"))
}

func TestRuns_DiscardHandler(t *testing.T) {
	run := &run.Run{ID: testutils.ParseID(t, "run-1")}
	h := &Handlers{
		Runs: &fakeRunClient{},
	}

	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
	w := httptest.NewRecorder()
	h.discardRun(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Equal(t, paths.Run(run.ID), w.Header().Get("HX-Redirect"))
}

func TestWebHandlers_CreateRun_Connected(t *testing.T) {
	run := &run.Run{ID: testutils.ParseID(t, "run-1")}
	h := &Handlers{
		Runs: &fakeRunClient{run: run},
	}

	q := "/?workspace_id=run-123&operation=plan-only&connected=true"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run(run.ID))
}

func TestWebHandlers_CreateRun_Unconnected(t *testing.T) {
	run := &run.Run{ID: testutils.ParseID(t, "run-1")}
	h := &Handlers{
		Runs: &fakeRunClient{run: run},
	}

	q := "/?workspace_id=run-123&operation=plan-only&connected=false"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run(run.ID))
}

func TestTailLogs(t *testing.T) {
	chunks := make(chan run.Chunk, 1)
	h := &Handlers{
		Runs: &fakeRunClient{
			run:    &run.Run{ID: testutils.ParseID(t, "run-1")},
			chunks: chunks,
		},
	}

	r := httptest.NewRequest("", "/?offset=0&phase=plan&run_id=run-123", nil)
	w := httptest.NewRecorder()

	// send one event and then close.
	chunks <- run.Chunk{Data: []byte("some logs")}
	close(chunks)

	done := make(chan struct{})
	go func() {
		h.tailRun(w, r)

		want := "data: {\"html\":\"some logs\\u003cbr\\u003e\",\"offset\":9}\nevent: log_update\n\ndata: no more logs\nevent: log_finished\n\n"
		assert.Equal(t, want, w.Body.String())

		done <- struct{}{}
	}()
	<-done
}

type fakeRunClient struct {
	RunService
	run    *run.Run
	chunks chan run.Chunk
}

func (f *fakeRunClient) List(_ context.Context, opts run.ListOptions) (*resource.Page[*run.Run], error) {
	return resource.NewPage([]*run.Run{f.run}, opts.PageOptions, nil), nil
}

func (f *fakeRunClient) Get(ctx context.Context, id resource.TfeID) (*run.Run, error) {
	return f.run, nil
}

func (f *fakeRunClient) GetChunk(ctx context.Context, opts run.GetChunkOptions) (run.Chunk, error) {
	return run.Chunk{}, nil
}

func (f *fakeRunClient) Cancel(ctx context.Context, id resource.TfeID) error {
	return nil
}

func (f *fakeRunClient) ForceCancel(ctx context.Context, id resource.TfeID) error {
	return nil
}

func (f *fakeRunClient) Discard(ctx context.Context, id resource.TfeID) error {
	return nil
}

func (f *fakeRunClient) Create(ctx context.Context, workspaceID resource.TfeID, opts run.CreateOptions) (*run.Run, error) {
	return f.run, nil
}

func (f *fakeRunClient) Tail(context.Context, run.TailOptions) (<-chan run.Chunk, error) {
	return f.chunks, nil
}

type fakeWorkspaceClient struct {
	WorkspaceService
	ws *workspace.Workspace
}

func (f *fakeWorkspaceClient) Get(context.Context, resource.TfeID) (*workspace.Workspace, error) {
	return f.ws, nil
}

type fakeConfigsClient struct{}

func (f *fakeConfigsClient) GetSourceIcon(source source.Source) templ.Component {
	return templ.Raw("")
}
