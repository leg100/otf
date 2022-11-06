package html

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestListWorkspacesHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	workspaces := []*otf.Workspace{
		otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{}),
		otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{}),
		otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{}),
		otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{}),
		otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{}),
	}
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{workspaces: workspaces})

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestListWorkspaceReposHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	repos := []*otf.Repo{
		otf.NewTestRepo(),
		otf.NewTestRepo(),
		otf.NewTestRepo(),
		otf.NewTestRepo(),
		otf.NewTestRepo(),
	}
	provider := otf.NewTestVCSProvider(org, otf.NewTestCloud(otf.WithRepos(repos...)))
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{provider: provider})

	q := "/?organization_name=fake-org&workspace_name=fake-workspace&vcs_provider_id=fake-provider"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSRepos(w, r)
	assert.Equal(t, 200, w.Code)

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", q+"&page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaceVCSRepos(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", q+"&page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaceVCSRepos(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", q+"&page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaceVCSRepos(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

type fakeWorkspaceHandlerApp struct {
	workspaces []*otf.Workspace
	provider   *otf.VCSProvider
	otf.Application
}

func (f *fakeWorkspaceHandlerApp) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

func (f *fakeWorkspaceHandlerApp) GetVCSProvider(ctx context.Context, providerID, organization string) (*otf.VCSProvider, error) {
	return f.provider, nil
}
