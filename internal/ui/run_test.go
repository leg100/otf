package ui

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestListRunsHandler(t *testing.T) {
	h := &runHandlers{
		workspaces: &fakeWorkspaceClient{
			ws: workspace.NewTestWorkspace(t, nil),
		},
		runs: &fakeRunClient{
			run: &run.Run{},
		},
		configs:    &fakeConfigsClient{},
		authorizer: authz.NewAllowAllAuthorizer(),
	}
	user := &user.User{ID: resource.NewTfeID(resource.UserKind)}

	r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page=1", nil)
	r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
	w := httptest.NewRecorder()
	h.list(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestWeb_GetHandler(t *testing.T) {
	h := &runHandlers{
		workspaces: &fakeWorkspaceClient{
			ws: workspace.NewTestWorkspace(t, nil),
		},
		runs: &fakeRunClient{
			run: &run.Run{
				Engine: engine.Default,
			},
		},
		configs:    &fakeConfigsClient{},
		authorizer: authz.NewAllowAllAuthorizer(),
	}
	r := httptest.NewRequest("GET", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

//
//func TestRuns_CancelHandler(t *testing.T) {
//	run := &Run{ID: testutils.ParseID(t, "run-1")}
//	h := newTestWebHandlers(t, withRuns(run))
//
//	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
//	w := httptest.NewRecorder()
//	h.cancel(w, r)
//	testutils.AssertRedirect(t, w, paths.Run(run.ID))
//}
//
//func TestWebHandlers_CreateRun_Connected(t *testing.T) {
//	run := &Run{ID: testutils.ParseID(t, "run-1")}
//	h := newTestWebHandlers(t,
//		withRuns(run),
//		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123"), Connection: &workspace.Connection{}}),
//	)
//
//	q := "/?workspace_id=run-123&operation=plan-only&connected=true"
//	r := httptest.NewRequest("POST", q, nil)
//	w := httptest.NewRecorder()
//	h.createRun(w, r)
//	testutils.AssertRedirect(t, w, paths.Run(run.ID))
//}
//
//func TestWebHandlers_CreateRun_Unconnected(t *testing.T) {
//	run := &Run{ID: testutils.ParseID(t, "run-1")}
//	h := newTestWebHandlers(t,
//		withRuns(run),
//	)
//
//	q := "/?workspace_id=run-123&operation=plan-only&connected=false"
//	r := httptest.NewRequest("POST", q, nil)
//	w := httptest.NewRecorder()
//	h.createRun(w, r)
//	testutils.AssertRedirect(t, w, paths.Run(run.ID))
//}
//
//func TestTailLogs(t *testing.T) {
//	chunks := make(chan Chunk, 1)
//	handlers := &webHandlers{
//		logger: logr.Discard(),
//		runs:   &fakeTailService{chunks: chunks},
//	}
//
//	r := httptest.NewRequest("", "/?offset=0&phase=plan&run_id=run-123", nil)
//	w := httptest.NewRecorder()
//
//	// send one event and then close.
//	chunks <- Chunk{Data: []byte("some logs")}
//	close(chunks)
//
//	done := make(chan struct{})
//	go func() {
//		handlers.tailRun(w, r)
//
//		// should receive base64 encoded event
//		want := "data: {\"html\":\"some logs\\u003cbr\\u003e\",\"offset\":9}\nevent: log_update\n\ndata: no more logs\nevent: log_finished\n\n"
//		assert.Equal(t, want, w.Body.String())
//
//		done <- struct{}{}
//	}()
//	<-done
//}
//
//type fakeTailService struct {
//	webRunClient
//
//	chunks chan Chunk
//}
//
//func (f *fakeTailService) Tail(context.Context, TailOptions) (<-chan Chunk, error) {
//	return f.chunks, nil
//}

type fakeRunClient struct {
	run *run.Run
	runClient
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

type fakeWorkspaceClient struct {
	ws *workspace.Workspace
	runWorkspaceClient
}

func (f *fakeWorkspaceClient) Get(context.Context, resource.TfeID) (*workspace.Workspace, error) {
	return f.ws, nil
}

type fakeConfigsClient struct{}

func (f *fakeConfigsClient) GetSourceIcon(source source.Source) templ.Component {
	return templ.Raw("")
}
