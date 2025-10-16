package ui

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestTeam_WebHandlers(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		h := &teamHandlers{}
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("create", func(t *testing.T) {
		h := &teamHandlers{}
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("update", func(t *testing.T) {
		team := &team.Team{Name: "acme-org", ID: testutils.ParseID(t, "team-123")}
		h := &teamHandlers{teams: &fakeTeamService{team: team}}

		q := "/?team_id=team-123&manage_workspaces=true"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		h.updateTeam(w, r)
		testutils.AssertRedirect(t, w, paths.Team(team.ID))
	})

	t.Run("list", func(t *testing.T) {
		team := &team.Team{Name: "acme-org", ID: testutils.ParseID(t, "team-123")}
		h := &teamHandlers{
			teams:      &fakeTeamService{team: team},
			authorizer: authz.NewAllowAllAuthorizer(),
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
		h := &teamHandlers{teams: &fakeTeamService{team: team}}
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
