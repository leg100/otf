package run

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestListRunsHandler(t *testing.T) {
	runs := make([]*Run, 201)
	for i := 1; i <= 201; i++ {
		runs[i-1] = &Run{ID: fmt.Sprintf("run-%d", i)}
	}
	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: "ws-123"}),
		withRuns(runs...),
	)

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=1", nil)
		r = r.WithContext(internal.AddSubjectToContext(r.Context(), &auth.User{ID: "janitor"}))
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=2", nil)
		r = r.WithContext(internal.AddSubjectToContext(r.Context(), &auth.User{ID: "janitor"}))
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=3", nil)
		r = r.WithContext(internal.AddSubjectToContext(r.Context(), &auth.User{ID: "janitor"}))
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
		withRuns((&Run{ID: "run-123", WorkspaceID: "ws-1"}).updateStatus(RunPending, nil)),
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

func TestWebHandlers_CreateRun_Connected(t *testing.T) {
	h := newTestWebHandlers(t,
		withRuns(&Run{ID: "run-1"}),
		withWorkspace(&workspace.Workspace{ID: "ws-123", Connection: &workspace.Connection{}}),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=true"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}

func TestWebHandlers_CreateRun_Unconnected(t *testing.T) {
	h := newTestWebHandlers(t,
		withRuns(&Run{ID: "run-1"}),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=false"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}
