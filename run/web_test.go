package run

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRunsHandler(t *testing.T) {
	app := newTestWebHandlers(t,
		withWorkspaces(&otf.Workspace{ID: "ws-123"}),
		withRuns(
			&otf.Run{ID: "run-1"},
			&otf.Run{ID: "run-2"},
			&otf.Run{ID: "run-3"},
			&otf.Run{ID: "run-4"},
			&otf.Run{ID: "run-5"},
		),
	)

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestRuns_CancelHandler(t *testing.T) {
	app := newTestWebHandlers(t, withRuns(&otf.Run{ID: "run-1", WorkspaceID: "ws-1"}))

	r := httptest.NewRequest("POST", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	app.cancel(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.Runs("ws-1"), redirect.Path)
	}
}

type (
	fakeWebServices struct {
		runs []*otf.Run
		ws   *otf.Workspace

		service

		otf.RunService
		otf.WorkspaceService
	}

	fakeWebServiceOption func(*fakeWebServices)
)

func withWorkspaces(workspace *otf.Workspace) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.ws = workspace
	}
}

func withRuns(runs ...*otf.Run) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.runs = runs
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
		svc:              &svc,
	}
}

func (f *fakeWebServices) GetWorkspaceByName(context.Context, string, string) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) list(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
}

func (f *fakeWebServices) get(ctx context.Context, runID string) (*otf.Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) cancel(ctx context.Context, runID string) error { return nil }
