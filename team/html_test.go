package team

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

// TestTeam_GetHandler tests the getTeam handler. The getTeam page renders
// permissions only if the authenticated user is an owner, so the test sets that
// up first.
func TestTeam_GetHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	owners := otf.NewTestOwners(t, org)
	owner := otf.NewTestUser(t, otf.WithTeamMemberships(owners))
	app := newFakeWebApp(t, &fakeTeamHandlerApp{
		team:    owners,
		members: []*otf.User{owner},
	})

	q := "/?team_id=team-123"
	r := httptest.NewRequest("GET", q, nil)
	r = r.WithContext(otf.AddSubjectToContext(r.Context(), owner))
	w := httptest.NewRecorder()
	app.getTeam(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestTeam_UpdateHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	team := otf.NewTestTeam(t, org)
	app := newFakeWebApp(t, &fakeTeamHandlerApp{
		team: team,
	})

	q := "/?team_id=team-123&manage_workspaces=true"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.updateTeam(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.Team(team.ID()), redirect.Path)
	}
}

func TestTeam_ListHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	team := otf.NewTestTeam(t, org)
	app := newFakeWebApp(t, &fakeTeamHandlerApp{
		team: team,
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listTeams(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

type fakeTeamHandlerApp struct {
	team    *otf.Team
	members []*otf.User

	otf.Application
}

func (f *fakeTeamHandlerApp) GetTeam(ctx context.Context, teamID string) (*otf.Team, error) {
	return f.team, nil
}

func (f *fakeTeamHandlerApp) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	return []*otf.Team{f.team}, nil
}

func (f *fakeTeamHandlerApp) UpdateTeam(ctx context.Context, teamID string, opts otf.UpdateTeamOptions) (*otf.Team, error) {
	f.team.Update(opts)
	return f.team, nil
}

func (f *fakeTeamHandlerApp) ListTeamMembers(ctx context.Context, teamID string) ([]*otf.User, error) {
	return f.members, nil
}
