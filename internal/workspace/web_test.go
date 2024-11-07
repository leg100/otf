package workspace

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestNewWorkspaceHandler(t *testing.T) {
	h := &webHandlers{Renderer: testutils.NewRenderer(t)}

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.newWorkspace(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWorkspace_Create(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-123")}
	h := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

	form := strings.NewReader(url.Values{
		"name": {"dev"},
	}.Encode())
	r := httptest.NewRequest("POST", "/?organization_name=acme-corp", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.createWorkspace(w, r)
	if assert.Equal(t, 302, w.Code, "output: %s", w.Body.String()) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspace(ws.ID.String()), redirect.Path)
	}
}

func TestGetWorkspaceHandler(t *testing.T) {
	privilegedUser = resource.NewID(resource.UserKind)

	tests := []struct {
		name      string
		workspace *Workspace
	}{
		{
			"unlocked", &Workspace{ID: resource.ParseID("ws-unlocked"), Lock: &lock{}},
		},
		{
			"locked", &Workspace{ID: resource.ParseID("ws-locked"), Lock: &lock{ID: &privilegedUser}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &webHandlers{
				Renderer: testutils.NewRenderer(t),
				client:   &FakeService{Workspaces: []*Workspace{tt.workspace}},
			}

			q := "/?workspace_id=ws-123"
			r := httptest.NewRequest("GET", q, nil)
			r = r.WithContext(authz.AddSubjectToContext(r.Context(), &user.User{ID: resource.ParseID("janitor")}))
			w := httptest.NewRecorder()
			app.getWorkspace(w, r)
			assert.Equal(t, 200, w.Code, w.Body.String())
		})
	}
}

func TestWorkspace_GetByName(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-123")}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

	q := "/?organization_name=acme-corp&workspace_name=fake-ws"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.getWorkspaceByName(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Workspace(ws.ID.String()), redirect.Path)
	}
}

func TestEditWorkspaceHandler(t *testing.T) {
	tests := []struct {
		name   string
		ws     *Workspace
		teams  []*team.Team
		policy authz.WorkspacePolicy
		user   user.User
		want   func(t *testing.T, doc *html.Node)
	}{
		{
			name: "default",
			ws:   &Workspace{ID: resource.ParseID("ws-123")},
			user: user.SiteAdmin,
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
			name: "with policy",
			ws:   &Workspace{ID: resource.ParseID("ws-123")},
			user: user.SiteAdmin,
			policy: authz.WorkspacePolicy{
				Permissions: []authz.WorkspacePermission{
					{TeamID: resource.ParseID("team-1"), Role: rbac.WorkspaceAdminRole},
					{TeamID: resource.ParseID("team-4"), Role: rbac.WorkspacePlanRole},
				},
			},
			teams: []*team.Team{
				{ID: resource.ParseID("team-1"), Name: "bosses"},
				{ID: resource.ParseID("team-2"), Name: "stewards"},
				{ID: resource.ParseID("team-3"), Name: "cleaners"},
				{ID: resource.ParseID("team-4"), Name: "workers"},
			},
			want: func(t *testing.T, doc *html.Node) {
				// tabulate existing assigned permissions
				findText(t, doc, "bosses", "//tr[@id='permissions-bosses']/td[1]")
				findText(t, doc, "admin", "//tr[@id='permissions-bosses']/td[2]//option[@selected]")

				findText(t, doc, "workers", "//tr[@id='permissions-workers']/td[1]")
				findText(t, doc, "plan", "//tr[@id='permissions-workers']/td[2]//option[@selected]")

				// form for assigning permissions to unassigned teams
				findText(t, doc, "stewards", "//select[@form='permissions-add-form']/option[@value='team-2']")
				findText(t, doc, "cleaners", "//select[@form='permissions-add-form']/option[@value='team-3']")

				// form should not include teams already assigned permissions,
				// nor owners team
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='team-1']")
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='team-4']")
				findTextNot(t, doc, "//select[@form='permissions-add-form']/option[@value='owners']")
			},
		},
		{
			name: "connected repo",
			ws:   &Workspace{ID: resource.ParseID("ws-123"), Connection: &Connection{Repo: "leg100/otf"}},
			user: user.SiteAdmin,
			want: func(t *testing.T, doc *html.Node) {
				got := htmlquery.FindOne(doc, "//button[@id='disconnect-workspace-repo-button']")
				assert.NotNil(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &webHandlers{
				Renderer: testutils.NewRenderer(t),
				client: &FakeService{
					Policy:     tt.policy,
					Workspaces: []*Workspace{tt.ws},
				},
				teams: &fakeTeamService{teams: tt.teams},
				vcsproviders: &fakeVCSProviderService{
					providers: []*vcsprovider.VCSProvider{{}},
				},
			}

			q := "/?workspace_id=ws-123"
			r := httptest.NewRequest("GET", q, nil)
			r = r.WithContext(authz.AddSubjectToContext(context.Background(), &tt.user))
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
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
		workspaces[i-1] = &Workspace{ID: resource.ParseID(fmt.Sprintf("ws-%d", i))}
	}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: workspaces},
	}

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=1", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=2", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?organization_name=acme&page[number]=3", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
		w := httptest.NewRecorder()
		app.listWorkspaces(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}

func TestListWorkspacesHandler_WithLatestRun(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-foo"), LatestRun: &LatestRun{Status: "applied", ID: resource.ParseID("run-123")}}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

	r := httptest.NewRequest("GET", "/?organization_name=acme", nil)
	r = r.WithContext(authz.AddSubjectToContext(context.Background(), &user.SiteAdmin))
	w := httptest.NewRecorder()
	app.listWorkspaces(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestDeleteWorkspace(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
		vcsproviders: &fakeVCSProviderService{
			providers: []*vcsprovider.VCSProvider{
				{},
				{},
				{},
			},
		},
	}

	q := "/?workspace_id=ws-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSProviders(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestListWorkspaceReposHandler(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
		vcsproviders: &fakeVCSProviderService{
			providers: []*vcsprovider.VCSProvider{
				{},
			},
			repos: []string{
				vcs.NewTestRepo(),
				vcs.NewTestRepo(),
				vcs.NewTestRepo(),
				vcs.NewTestRepo(),
				vcs.NewTestRepo(),
			},
		},
	}

	q := "/?workspace_id=ws-123&vcs_provider_id=fake-provider"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listWorkspaceVCSRepos(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestConnectWorkspaceHandler(t *testing.T) {
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
	ws := &Workspace{ID: resource.ParseID("ws-123"), Organization: "acme-corp"}
	app := &webHandlers{
		Renderer: testutils.NewRenderer(t),
		client:   &FakeService{Workspaces: []*Workspace{ws}},
	}

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
	policy := authz.WorkspacePolicy{Permissions: []authz.WorkspacePermission{
		{TeamID: resource.ParseID("bosses"), Role: rbac.WorkspaceAdminRole},
		{TeamID: resource.ParseID("workers"), Role: rbac.WorkspacePlanRole},
	}}
	teams := []*team.Team{
		{ID: resource.ParseID("owners")},
		{ID: resource.ParseID("bosses")},
		{ID: resource.ParseID("stewards")},
		{ID: resource.ParseID("cleaners")},
		{ID: resource.ParseID("workers")},
	}
	want := []*team.Team{
		{ID: resource.ParseID("owners")},
		{ID: resource.ParseID("stewards")},
		{ID: resource.ParseID("cleaners")},
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
