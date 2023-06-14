package run

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
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

func TestWeb_GetHandler(t *testing.T) {
	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: "ws-123"}),
		withRuns(&Run{ID: "run-123", WorkspaceID: "ws-1"}),
	)

	r := httptest.NewRequest("GET", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, "output: %s", w.Body.String())
}

func TestRuns_CancelHandler(t *testing.T) {
	h := newTestWebHandlers(t, withRuns(&Run{ID: "run-1", WorkspaceID: "ws-1"}))

	r := httptest.NewRequest("POST", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.cancel(w, r)
	testutils.AssertRedirect(t, w, paths.Runs("ws-1"))
}

func TestWebHandlers_StartRun(t *testing.T) {
	run := &Run{ID: "run-1"}
	h := newTestWebHandlers(t, withRuns(run))

	q := "/?workspace_id=run-123&operation=plan-only"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.startRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}
