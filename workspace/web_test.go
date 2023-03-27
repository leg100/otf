package workspace

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: rename tests to TestWorkspace_<handler>

func TestNewWorkspaceHandler(t *testing.T) {
	app := fakeWebHandlers(t)

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newWorkspace(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestWorkspace_Create(t *testing.T) {
	ws := &Workspace{ID: "ws-123"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	form := strings.NewReader(url.Values{
		"name": {"dev"},
	}.Encode())
	r := httptest.NewRequest("POST", "/?organization_name=acme-corp", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.createWorkspace(w, r)
	if assert.Equal(t, 302, w.Code, "output: %s", w.Body.String()) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspace(ws.ID), redirect.Path)
	}
}

func TestGetWorkspaceHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.getWorkspace(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}

	// TODO: another test for retrieving latest run
}

func TestWorkspace_GetByName(t *testing.T) {
	ws := &Workspace{ID: "ws-123"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	q := "/?organization_name=acme-corp&workspace_name=fake-ws"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.getWorkspaceByName(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspace(ws.ID), redirect.Path)
	}
}

func TestEditWorkspaceHandler(t *testing.T) {
	tests := []struct {
		name   string
		ws     *Workspace
		teams  []*auth.Team
		policy otf.WorkspacePolicy
	}{
		{
			name: "default",
			ws:   &Workspace{ID: "ws-123"},
		},
		{
			name: "with policy",
			ws:   &Workspace{ID: "ws-123"},
			policy: otf.WorkspacePolicy{
				Permissions: []otf.WorkspacePermission{
					{
						Team: "devs",
						Role: rbac.WorkspaceAdminRole,
					},
				},
			},
		},
		{
			name:  "with unassigned teams",
			ws:    &Workspace{ID: "ws-123"},
			teams: []*auth.Team{{Name: "devs"}},
		},
		{
			name: "connected repo",
			ws:   &Workspace{ID: "ws-123", Connection: &repo.Connection{Repo: "leg100/otf"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fakeWebHandlers(t,
				withWorkspaces(tt.ws), withTeams(tt.teams...), withPolicy(tt.policy))

			q := "/?workspace_id=ws-123"
			r := httptest.NewRequest("GET", q, nil)
			w := httptest.NewRecorder()
			app.editWorkspace(w, r)
			assert.Equal(t, 200, w.Code, "output: %s", w.Body.String())
		})
	}
}

func TestUpdateWorkspaceHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	form := strings.NewReader(url.Values{
		"workspace_id": {"ws-123"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.updateWorkspace(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/workspaces/ws-123/edit", redirect.Path)
	}
}

func TestListWorkspacesHandler(t *testing.T) {
	ws1 := &Workspace{ID: "ws-1"}
	ws2 := &Workspace{ID: "ws-2"}
	ws3 := &Workspace{ID: "ws-3"}
	ws4 := &Workspace{ID: "ws-4"}
	ws5 := &Workspace{ID: "ws-5"}
	app := fakeWebHandlers(t, withWorkspaces(ws1, ws2, ws3, ws4, ws5))

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestDeleteWorkspace(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.deleteWorkspace(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspaces("acme-corp"), redirect.Path)
	}
}

func TestLockWorkspace(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	form := strings.NewReader(url.Values{
		"workspace_id": {"ws-123"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.lockWorkspace(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/workspaces/ws-123", redirect.Path)
	}
}

func TestUnlockWorkspace(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	form := strings.NewReader(url.Values{
		"workspace_id": {"ws-123"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.unlockWorkspace(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/workspaces/ws-123", redirect.Path)
	}
}

func TestListWorkspaceProvidersHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws), withVCSProviders(
		&vcsprovider.VCSProvider{},
		&vcsprovider.VCSProvider{},
		&vcsprovider.VCSProvider{},
	))

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSProviders(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestListWorkspaceReposHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws), withVCSProviders(&vcsprovider.VCSProvider{}),
		withRepos(
			cloud.NewTestRepo(),
			cloud.NewTestRepo(),
			cloud.NewTestRepo(),
			cloud.NewTestRepo(),
			cloud.NewTestRepo(),
		))

	q := "/?workspace_id=ws-123&vcs_provider_id=fake-provider"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSRepos(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestConnectWorkspaceHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws), withVCSProviders(&vcsprovider.VCSProvider{}))

	form := strings.NewReader(url.Values{
		"workspace_id":    {"ws-123"},
		"vcs_provider_id": {"fake-provider"},
		"identifier":      {"acme/myrepo"},
		"http_url":        {"https://fake-cloud/acme/myrepo"},
		"branch":          {"master"},
		"cloud":           {"fake-cloud"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.connect(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/workspaces/ws-123", redirect.Path)
	}
}

func TestDisconnectWorkspaceHandler(t *testing.T) {
	ws := &Workspace{ID: "ws-123", Organization: "acme-corp"}
	app := fakeWebHandlers(t, withWorkspaces(ws))

	form := strings.NewReader(url.Values{
		"workspace_id": {"ws-123"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.disconnect(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/workspaces/ws-123", redirect.Path)
	}
}

type (
	fakeWebService struct {
		workspaces []*Workspace
		providers  []*vcsprovider.VCSProvider
		repos      []string
		policy     otf.WorkspacePolicy
		teams      []*auth.Team

		Service

		auth.TeamService
		VCSProviderService
	}

	fakeWebServiceOption func(*fakeWebService)
)

func withWorkspaces(workspaces ...*Workspace) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.workspaces = workspaces
	}
}

func withVCSProviders(providers ...*vcsprovider.VCSProvider) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.providers = providers
	}
}

func withRepos(repos ...string) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.repos = repos
	}
}

func withPolicy(policy otf.WorkspacePolicy) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.policy = policy
	}
}

func withTeams(teams ...*auth.Team) fakeWebServiceOption {
	return func(svc *fakeWebService) {
		svc.teams = teams
	}
}

func fakeWebHandlers(t *testing.T, opts ...fakeWebServiceOption) *webHandlers {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)

	var svc fakeWebService
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		Renderer:           renderer,
		TeamService:        &svc,
		VCSProviderService: &svc,
		svc:                &svc,
	}
}

func (f *fakeWebService) GetVCSProvider(ctx context.Context, providerID string) (*vcsprovider.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeWebService) ListVCSProviders(context.Context, string) ([]*vcsprovider.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeWebService) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeWebService) GetPolicy(context.Context, string) (otf.WorkspacePolicy, error) {
	return f.policy, nil
}

func (f *fakeWebService) ListTeams(context.Context, string) ([]*auth.Team, error) {
	return f.teams, nil
}

func (f *fakeWebService) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	return &fakeWebCloudClient{repos: f.repos}, nil
}

func (f *fakeWebService) CreateWorkspace(context.Context, CreateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) UpdateWorkspace(context.Context, string, UpdateOptions) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) ListWorkspaces(ctx context.Context, opts ListOptions) (*WorkspaceList, error) {
	return &WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

func (f *fakeWebService) GetWorkspace(context.Context, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) GetWorkspaceByName(context.Context, string, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) DeleteWorkspace(context.Context, string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) LockWorkspace(context.Context, string, *string) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) UnlockWorkspace(context.Context, string, *string, bool) (*Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWebService) connect(context.Context, string, ConnectOptions) (*repo.Connection, error) {
	return nil, nil
}

func (f *fakeWebService) disconnect(context.Context, string) error {
	return nil
}

type fakeWebCloudClient struct {
	repos []string

	cloud.Client
}

func (f *fakeWebCloudClient) ListRepositories(ctx context.Context, opts cloud.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}
