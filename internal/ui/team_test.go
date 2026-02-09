package ui

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_WebHandlers(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		h := &Handlers{}
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("create", func(t *testing.T) {
		h := &Handlers{}
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("update", func(t *testing.T) {
		team := &team.Team{Name: "acme-org", ID: testutils.ParseID(t, "team-123")}
		h := &Handlers{Teams: &fakeTeamService{team: team}}

		q := "/?team_id=team-123&manage_workspaces=true"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.updateTeam(w, r)
		testutils.AssertRedirect(t, w, paths.Team(team.ID))
	})

	t.Run("get", func(t *testing.T) {
		org1 := organization.NewTestName(t)
		// page only renders successfully if authenticated user is an owner.
		owners := &team.Team{Name: "owners", Organization: org1}
		owner, err := user.NewUser(uuid.NewString(), user.WithTeams(owners))
		require.NoError(t, err)
		h := &Handlers{
			Authorizer: authz.NewAllowAllAuthorizer(),
			Teams:      &fakeTeamService{team: owners},
			Users:      &fakeUserService{user: owner},
		}

		q := "/?team_id=team-123"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.getTeam(w, r)
		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		team := &team.Team{Name: "acme-org", ID: testutils.ParseID(t, "team-123")}
		h := &Handlers{
			Teams:      &fakeTeamService{team: team},
			Authorizer: authz.NewAllowAllAuthorizer(),
		}
		// make request with user with full perms, to ensure parts of
		// page that are hidden to unprivileged users are not hidden.
		userCtx := authz.AddSubjectToContext(context.Background(), &authz.Superuser{})

		q := "/?organization_name=acme"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(userCtx)
		w := httptest.NewRecorder()
		h.listTeams(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("delete", func(t *testing.T) {
		team := &team.Team{Name: "acme-org", ID: testutils.ParseID(t, "team-123"), Organization: organization.NewTestName(t)}
		h := &Handlers{Teams: &fakeTeamService{team: team}}
		q := "/?team_id=team-123"
		r := httptest.NewRequest("POST", q, nil)
		w := httptest.NewRecorder()
		h.deleteTeam(w, r)
		testutils.AssertRedirect(t, w, paths.Teams(team.Organization))
	})
}

type fakeTeamService struct {
	team *team.Team
}

func (f *fakeTeamService) Create(context.Context, organization.Name, team.CreateTeamOptions) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) Update(context.Context, resource.TfeID, team.UpdateTeamOptions) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) Get(context.Context, organization.Name, string) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) GetByID(context.Context, resource.TfeID) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) List(context.Context, organization.Name) ([]*team.Team, error) {
	return []*team.Team{f.team}, nil
}

func (f *fakeTeamService) Delete(context.Context, resource.TfeID) error {
	return nil
}
