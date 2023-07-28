package workspace

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
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
	tests := []struct {
		name      string
		workspace *Workspace
	}{
		{
			"unlocked", &Workspace{ID: "ws-unlocked"},
		},
		{
			"locked", &Workspace{ID: "ws-locked", Lock: &Lock{id: "janitor", LockKind: UserLock}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fakeWebHandlers(t, withWorkspaces(tt.workspace))

			q := "/?workspace_id=ws-123"
			r := httptest.NewRequest("GET", q, nil)
			r = r.WithContext(internal.AddSubjectToContext(r.Context(), &auth.User{ID: "janitor"}))
			w := httptest.NewRecorder()
			app.getWorkspace(w, r)
			assert.Equal(t, 200, w.Code, w.Body.String())
		})
	}
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
		policy internal.WorkspacePolicy
		user   auth.User
		want   func(t *testing.T, doc *html.Node)
	}{
		{
			name: "default",
			ws:   &Workspace{ID: "ws-123"},
			user: auth.SiteAdmin,
			want: func(t *testing.T, doc *html.Node) {
				// always show built-in owners permission
				findText(t, doc, "owners", "//div[@id='permissions-container']//tbody//tr[1]/td[1]")
				findText(t, doc, "admin", "//div[@id='permissions-container']//tbody//tr[1]/td[2]")

				// all buttons should be enabled
				buttons := htmlquery.Find(doc, `//button`)
				for _, btn := range buttons {
					assert.NotContains(t, testutils.AttrMap(btn), "disabled")
				}
			},
		},
		{
			name: "insufficient privileges",
			ws:   &Workspace{ID: "ws-123"},
			user: auth.User{}, // user with no privs
			want: func(t *testing.T, doc *html.Node) {
				// all buttons should be disabled
				buttons := htmlquery.Find(doc, `//button`)
				for _, btn := range buttons {
					assert.Contains(t, testutils.AttrMap(btn), "disabled")
				}
			},
		},
		{
			name: "with policy",
			ws:   &Workspace{ID: "ws-123"},
			user: auth.SiteAdmin,
			policy: internal.WorkspacePolicy{
				Permissions: []internal.WorkspacePermission{
					{Team: "bosses", Role: rbac.WorkspaceAdminRole},
					{Team: "workers", Role: rbac.WorkspacePlanRole},
				},
			},
			teams: []*auth.Team{
				{Name: "bosses"},
				{Name: "stewards"},
				{Name: "cleaners"},
				{Name: "workers"},
			},
			want: func(t *testing.T, doc *html.Node) {
				// tabulate existing assigned permissions
				findText(t, doc, "bosses", "//tr[@id='permissions-bosses']/td[1]")
				findText(t, doc, "admin", "//tr[@id='permissions-bosses']/td[2]//option[@selected]")

				findText(t, doc, "workers", "//tr[@id='permissions-workers']/td[1]")
				findText(t, doc, "plan", "//tr[@id='permissions-workers']/td[2]//option[@selected]")

				// form for assigning permissions to unassigned teams
				findText(t, doc, "stewards", "//select[@form='permissions-add-form']/option[@value='stewards']")
				findText(t, doc, "cleaners", "//select[@form='permissions-add-form']/option[@value='cleaners']")

				// form should not include teams already assigned permissions,
				// nor owners team
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='bosses']")
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='workers']")
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='owners']")
			},
		},
		{
			name: "connected repo",
			ws:   &Workspace{ID: "ws-123", Connection: &Connection{Connection: &repo.Connection{Repo: "leg100/otf"}}},
			user: auth.SiteAdmin,
			want: func(t *testing.T, doc *html.Node) {
				got := htmlquery.FindOne(doc, "//button[@id='disconnect-workspace-repo-button']")
				assert.NotNil(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fakeWebHandlers(t,
				withWorkspaces(tt.ws),
				withTeams(tt.teams...),
				withPolicy(tt.policy),
				withVCSProviders(&vcsprovider.VCSProvider{}),
			)

			q := "/?workspace_id=ws-123"
			r := httptest.NewRequest("GET", q, nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &tt.user))
			w := httptest.NewRecorder()
			app.editWorkspace(w, r)
			assert.Equal(t, 200, w.Code, w.Body.String())

			if tt.want != nil {
				doc, err := htmlquery.Parse(w.Body)
				require.NoError(t, err)
				tt.want(t, doc)
			}
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
	workspaces := make([]*Workspace, 201)
	for i := 1; i <= 201; i++ {
		workspaces[i-1] = &Workspace{ID: fmt.Sprintf("ws-%d", i)}
	}
	app := fakeWebHandlers(t,
		withWorkspaces(workspaces...),
	)

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=1", nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin))
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=2", nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin))
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=3", nil)
		r = r.WithContext(internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin))
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
	assert.Equal(t, 200, w.Code, w.Body.String())
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

func TestFilterUnassigned(t *testing.T) {
	policy := internal.WorkspacePolicy{Permissions: []internal.WorkspacePermission{
		{Team: "bosses", Role: rbac.WorkspaceAdminRole},
		{Team: "workers", Role: rbac.WorkspacePlanRole},
	}}
	teams := []*auth.Team{
		{Name: "owners"},
		{Name: "bosses"},
		{Name: "stewards"},
		{Name: "cleaners"},
		{Name: "workers"},
	}
	want := []*auth.Team{
		{Name: "stewards"},
		{Name: "cleaners"},
	}
	got := filterUnassigned(policy, teams)
	assert.Equal(t, want, got)
}

func findText(t *testing.T, doc *html.Node, want, selector string) {
	t.Helper()

	got := htmlquery.FindOne(doc, selector)
	if assert.NotNil(t, got) {
		assert.Equal(t, want, htmlquery.InnerText(got))
	}
}

func findTextNot(t *testing.T, doc *html.Node, selector string) {
	got := htmlquery.FindOne(doc, selector)
	assert.Nil(t, got)
}
