package run

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestListRunsHandler(t *testing.T) {
	// NOTE: We can't easily unit test this handler because a
	// websocket is responsible for fetching the listing. Instead we rely on
	// integration tests.
	t.Skip()

	h := newTestWebHandlers(t,
		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123")}),
	)
	user := &user.User{ID: resource.NewTfeID(resource.UserKind)}

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page=1", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.listByOrganization(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page=2", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.listByOrganization(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?workspace_id=ws-123&page=3", nil)
		r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
		w := httptest.NewRecorder()
		h.listByOrganization(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestWeb_GetHandler(t *testing.T) {
	ws1 := workspace.NewTestWorkspace(t, nil)
	run1 := newTestRun(t, context.Background(), CreateOptions{})
	h := newTestWebHandlers(t,
		withWorkspace(ws1),
		withRuns(run1),
	)

	r := httptest.NewRequest("GET", "/?run_id=run-123", nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, "output: %s", w.Body.String())
}

func TestRuns_CancelHandler(t *testing.T) {
	run := &Run{ID: testutils.ParseID(t, "run-1")}
	h := newTestWebHandlers(t, withRuns(run))

	r := httptest.NewRequest("POST", "/?run_id=run-1", nil)
	w := httptest.NewRecorder()
	h.cancel(w, r)
	testutils.AssertRedirect(t, w, paths.Run(run.ID))
}

func TestWebHandlers_CreateRun_Connected(t *testing.T) {
	run := &Run{ID: testutils.ParseID(t, "run-1")}
	h := newTestWebHandlers(t,
		withRuns(run),
		withWorkspace(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123"), Connection: &workspace.Connection{}}),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=true"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run(run.ID))
}

func TestWebHandlers_CreateRun_Unconnected(t *testing.T) {
	run := &Run{ID: testutils.ParseID(t, "run-1")}
	h := newTestWebHandlers(t,
		withRuns(run),
	)

	q := "/?workspace_id=run-123&operation=plan-only&connected=false"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	h.createRun(w, r)
	testutils.AssertRedirect(t, w, paths.Run(run.ID))
}
