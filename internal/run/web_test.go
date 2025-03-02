package run

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestListRunsHandler(t *testing.T) {
	runs := make([]*Run, 201)
	for i := 1; i <= 201; i++ {
		runs[i-1] = &Run{ID: testutils.ParseID(t, fmt.Sprintf("run-%d", i))}
	}
	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123")}),
		withRuns(runs...),
	)
	user := &user.User{ID: resource.NewID(resource.UserKind)}

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=1", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=2", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page[number]=3", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.list(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestWeb_GetHandler(t *testing.T) {
	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123")}),
		withRuns((&Run{ID: testutils.ParseID(t, "run-123"), WorkspaceID: testutils.ParseID(t, "ws-1")}).updateStatus(runstatus.Pending, nil)),
	)

	r := httptest.NewRequest("GET", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, "output: %s", w.Body.String())
}

func TestRuns_CancelHandler(t *testing.T) {
	h := newTestWebHandlers(t, withRuns(&Run{ID: testutils.ParseID(t, "run-1")}))

	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
	w := httptest.NewRecorder()
	h.cancel(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}

func TestWebHandlers_CreateRun_Connected(t *testing.T) {
	h := newTestWebHandlers(t,
		withRuns(&Run{ID: testutils.ParseID(t, "run-1")}),
		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123"), Connection: &workspace.Connection{}}),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=true"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}

func TestWebHandlers_CreateRun_Unconnected(t *testing.T) {
	h := newTestWebHandlers(t,
		withRuns(&Run{ID: testutils.ParseID(t, "run-1")}),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=false"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run("run-1"))
}
