package html

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: rename tests to TestWorkspace_<handler>

func TestNewWorkspaceHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{org: org})

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newWorkspace(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestGetWorkspaceHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{workspaces: []*otf.Workspace{ws}})

	q := "/?organization_name=acme-corp&workspace_name=fake-ws"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.getWorkspace(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestEditWorkspaceHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{workspaces: []*otf.Workspace{ws}})

	q := "/?"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.editWorkspace(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestListWorkspacesHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	workspaces := []*otf.Workspace{
		otf.NewTestWorkspace(t, org),
		otf.NewTestWorkspace(t, org),
		otf.NewTestWorkspace(t, org),
		otf.NewTestWorkspace(t, org),
		otf.NewTestWorkspace(t, org),
	}
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{org: org, workspaces: workspaces})

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
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{workspaces: []*otf.Workspace{ws}})

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.deleteWorkspace(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspaces(org.Name()), redirect.Path)
	}
}

func TestListWorkspaceProvidersHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	workspaces := []*otf.Workspace{
		otf.NewTestWorkspace(t, org),
	}
	providers := []*otf.VCSProvider{
		otf.NewTestVCSProvider(t, org),
		otf.NewTestVCSProvider(t, org),
		otf.NewTestVCSProvider(t, org),
	}
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		providers:  providers,
		workspaces: workspaces,
	})

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSProviders(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestListWorkspaceReposHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		providers: []*otf.VCSProvider{
			otf.NewTestVCSProvider(t, org),
		},
		workspaces: []*otf.Workspace{
			otf.NewTestWorkspace(t, org),
		},
		repos: []*otf.Repo{
			otf.NewTestRepo(),
			otf.NewTestRepo(),
			otf.NewTestRepo(),
			otf.NewTestRepo(),
			otf.NewTestRepo(),
		},
	})

	q := "/?workspace_id=ws-123&vcs_provider_id=fake-provider"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSRepos(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}

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

func TestConnectWorkspaceHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		workspaces: []*otf.Workspace{ws},
		providers:  []*otf.VCSProvider{otf.NewTestVCSProvider(t, org)},
	})

	form := strings.NewReader(url.Values{
		"organization_name": {"fake-org"},
		"workspace_name":    {"fake-workspace"},
		"vcs_provider_id":   {"fake-provider"},
		"identifier":        {"acme/myrepo"},
		"http_url":          {"https://fake-cloud/acme/myrepo"},
		"branch":            {"master"},
		"cloud":             {"fake-cloud"},
	}.Encode())
	r := httptest.NewRequest("POST", "/", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.connectWorkspace(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/workspaces/%s", ws.ID()), redirect.Path)
	}
}

func TestDisconnectWorkspaceHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
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
	app.disconnectWorkspace(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/workspaces/%s", ws.ID()), redirect.Path)
	}
}

func TestStartRunHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)
	cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	app := newFakeWebApp(t, &fakeWorkspaceHandlerApp{
		workspaces: []*otf.Workspace{ws},
		runs:       []*otf.Run{run},
	})

	q := "/?workspace_id=ws-123&strategy=plan-only"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	app.startRun(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("/runs/%s", run.ID()), redirect.Path)
	}
}

type fakeWorkspaceHandlerApp struct {
	org            *otf.Organization
	runs           []*otf.Run
	workspaces     []*otf.Workspace
	configVersions []*otf.ConfigurationVersion
	providers      []*otf.VCSProvider
	repos          []*otf.Repo

	otf.Application
}

func (f *fakeWorkspaceHandlerApp) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	return f.org, nil
}

func (f *fakeWorkspaceHandlerApp) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) DeleteWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

func (f *fakeWorkspaceHandlerApp) GetVCSProvider(ctx context.Context, providerID string) (*otf.VCSProvider, error) {
	return f.providers[0], nil
}

func (f *fakeWorkspaceHandlerApp) ListVCSProviders(context.Context, string) ([]*otf.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeWorkspaceHandlerApp) ConnectWorkspace(context.Context, otf.WorkspaceSpec, otf.ConnectWorkspaceOptions) (*otf.Workspace, error) {
	return f.workspaces[0], nil
}

func (f *fakeWorkspaceHandlerApp) DisconnectWorkspace(context.Context, otf.WorkspaceSpec) (*otf.Workspace, error) {
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

func (f *fakeWorkspaceHandlerApp) StartRun(context.Context, otf.WorkspaceSpec, otf.ConfigurationVersionCreateOptions) (*otf.Run, error) {
	return f.runs[0], nil
}

func (f *fakeWorkspaceHandlerApp) ListWorkspacePermissions(ctx context.Context, spec otf.WorkspaceSpec) ([]*otf.WorkspacePermission, error) {
	return nil, nil
}

func (f *fakeWorkspaceHandlerApp) ListTeams(context.Context, string) ([]*otf.Team, error) {
	return nil, nil
}

func (f *fakeWorkspaceHandlerApp) ListRepositories(ctx context.Context, providerID string, opts otf.ListOptions) (*otf.RepoList, error) {
	return &otf.RepoList{
		Items:      f.repos,
		Pagination: otf.NewPagination(opts, len(f.repos)),
	}, nil
}
