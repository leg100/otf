package run

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRunsHandler(t *testing.T) {
	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: "ws-123"}),
		withRuns(
			&Run{ID: "run-1"},
			&Run{ID: "run-2"},
			&Run{ID: "run-3"},
			&Run{ID: "run-4"},
			&Run{ID: "run-5"},
		),
	)

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestRuns_CancelHandler(t *testing.T) {
	h := newTestWebHandlers(t, withRuns(&Run{ID: "run-1", WorkspaceID: "ws-1"}))

	r := httptest.NewRequest("POST", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.cancel(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.Runs("ws-1"), redirect.Path)
	}
}

func TestWebHandlers_StartRun(t *testing.T) {
	tests := []struct {
		strategy        string
		wantSpeculative bool
	}{
		{"plan-only", true},
		{"plan-and-apply", false},
	}
	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			run := &Run{ID: "run-1"}
			h := newTestWebHandlers(t, withRuns(run))

			q := "/?workspace_id=run-123&strategy=" + tt.strategy
			r := httptest.NewRequest("POST", q, nil)
			w := httptest.NewRecorder()
			h.startRun(w, r)
			if assert.Equal(t, 302, w.Code) {
				redirect, _ := w.Result().Location()
				assert.Equal(t, paths.Run("run-1"), redirect.Path)
			}
			assert.Equal(t, tt.wantSpeculative, run.Speculative)
		})
	}
}

type (
	fakeWebServices struct {
		runs       []*Run
		ws         *workspace.Workspace
		gotOptions *otf.ConfigurationVersionCreateOptions

		service

		RunService
		workspace.WorkspaceService
	}

	fakeWebServiceOption func(*fakeWebServices)
)

func withWorkspace(workspace *workspace.Workspace) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.ws = workspace
	}
}

func withRuns(runs ...*Run) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.runs = runs
	}
}

func withServices(with *fakeWebServices) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		*svc = *with
	}
}

func newTestWebHandlers(t *testing.T, opts ...fakeWebServiceOption) *webHandlers {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)

	var svc fakeWebServices
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		Renderer:         renderer,
		WorkspaceService: &svc,
		starter:          &svc,
		svc:              &svc,
	}
}

func (f *fakeWebServices) GetWorkspaceByName(context.Context, string, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) list(ctx context.Context, opts RunListOptions) (*RunList, error) {
	return &RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
}

func (f *fakeWebServices) get(ctx context.Context, runID string) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) startRun(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*Run, error) {
	f.runs[0].Speculative = *opts.Speculative
	return f.runs[0], nil
}

func (f *fakeWebServices) cancel(ctx context.Context, runID string) error { return nil }
