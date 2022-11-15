package html

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestListWorkspaceProvidersHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	providers := []*otf.VCSProvider{
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
	}
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{providers: providers})

	q := "/?"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSProviders(w, r)
	assert.Equal(t, 200, w.Code)
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
	provider := otf.NewTestVCSProvider(t, org, otf.NewTestCloud(otf.WithRepos(repos...)))
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{providers: []*otf.VCSProvider{provider}})

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

func TestConnectWorkspaceRepoHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	repo := otf.NewTestRepo()
	provider := otf.NewTestVCSProvider(t, org, otf.NewTestCloud(otf.WithRepos(repo)))
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		workspaces: []*otf.Workspace{ws},
		providers:  []*otf.VCSProvider{provider},
	})

	form := strings.NewReader(url.Values{
		"organization_name": {"fake-org"},
		"workspace_name":    {"fake-workspace"},
		"vcs_provider_id":   {"fake-provider"},
		"identifier":        {"acme/myrepo"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.connectWorkspaceRepo(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/organizations/%s/workspaces/%s", org.Name(), ws.Name()), redirect.Path)
	}
}

func TestDisconnectWorkspaceRepoHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		workspaces: []*otf.Workspace{ws},
	})

	form := strings.NewReader(url.Values{
		"organization_name": {"fake-org"},
		"workspace_name":    {"fake-workspace"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.disconnectWorkspaceRepo(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/organizations/%s/workspaces/%s", org.Name(), ws.Name()), redirect.Path)
	}
}

func TestStartRunHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		runs:           []*otf.Run{run},
		workspaces:     []*otf.Workspace{ws},
		configVersions: []*otf.ConfigurationVersion{cv},
	})

	form := strings.NewReader(url.Values{
		"organization_name": {"fake-org"},
		"workspace_name":    {"fake-workspace"},
		"strategy":          {"plan-only"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.startRun(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/organizations/%s/workspaces/%s/runs/%s", org.Name(), ws.Name(), run.ID()), redirect.Path)
	}
}

func TestStartRun(t *testing.T) {
	ctx := context.Background()
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org, otf.NewTestCloud())

	t.Run("not connected to repo", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
		cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
		want := otf.NewRun(cv, ws, otf.RunCreateOptions{})
		app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
			runs:           []*otf.Run{want},
			workspaces:     []*otf.Workspace{ws},
			configVersions: []*otf.ConfigurationVersion{cv},
		})

		got, err := startRun(ctx, app, ws.SpecName(), false)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("connected to repo", func(t *testing.T) {
		repo := otf.NewTestVCSRepo(provider)
		ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{
			VCSRepo: repo,
		})
		cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
		want := otf.NewRun(cv, ws, otf.RunCreateOptions{})
		app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
			runs:           []*otf.Run{want},
			workspaces:     []*otf.Workspace{ws},
			configVersions: []*otf.ConfigurationVersion{cv},
		})

		got, err := startRun(ctx, app, ws.SpecName(), false)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

type fakeWorkspaceHandlerApp struct {
	runs           []*otf.Run
	workspaces     []*otf.Workspace
	configVersions []*otf.ConfigurationVersion
	providers      []*otf.VCSProvider
	otf.Application
}

func (f *fakeWorkspaceHandlerApp) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

func (f *fakeWorkspaceHandlerApp) GetVCSProvider(ctx context.Context, providerID, organization string) (*otf.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeWorkspaceHandlerApp) ListVCSProviders(context.Context, string) ([]*otf.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeWorkspaceHandlerApp) ConnectWorkspaceRepo(context.Context, otf.WorkspaceSpec, otf.VCSRepo) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) DisconnectWorkspaceRepo(context.Context, otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) UploadConfig(context.Context, string, []byte) error {
	return nil
}

func (f *fakeWorkspaceHandlerApp) GetLatestConfigurationVersion(context.Context, string) (*otf.ConfigurationVersion, error) {
	return f.configVersions[0], nil
}

func (f *fakeWorkspaceHandlerApp) CreateConfigurationVersion(context.Context, string, otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	return f.configVersions[0], nil
}

func (f *fakeWorkspaceHandlerApp) CloneConfigurationVersion(context.Context, string, otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	return f.configVersions[0], nil
}

func (f *fakeWorkspaceHandlerApp) CreateRun(context.Context, otf.WorkspaceSpec, otf.RunCreateOptions) (*otf.Run, error) {
	return f.runs[0], nil
}
