package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

// TestTeam_GetHandler tests the getTeam handler. The getTeam page renders
// permissions only if the authenticated user is an owner, so the test sets that
// up first.
func TestTeam_GetHandler(t *testing.T) {
	owners := newTestOwners(t, "acme-org")
	owner := NewUser(uuid.NewString(), WithTeams(owners))
	app := newFakeWeb(t, &fakeService{
		TeamService: &fakeTeamApp{
			team:    owners,
			members: []*User{owner},
		},
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
	team := NewTestTeam(t, "acme-org")
	app := newFakeWeb(t, &fakeService{
		TeamService: &fakeTeamApp{
			team: team,
		},
	})

	q := "/?team_id=team-123&manage_workspaces=true"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.updateTeam(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.Team(team.ID), redirect.Path)
	}
}

func TestTeam_ListHandler(t *testing.T) {
	team := NewTestTeam(t, "acme-org")
	app := newFakeWeb(t, &fakeService{
		TeamService: &fakeTeamApp{
			team: team,
		},
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listTeams(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}
