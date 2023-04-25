package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

func TestTeam_WebHandlers(t *testing.T) {
	userCtx := otf.AddSubjectToContext(context.Background(), &User{})

	t.Run("new", func(t *testing.T) {
		app := newFakeWeb(t, &fakeService{})
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		app.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("create", func(t *testing.T) {
		app := newFakeWeb(t, &fakeService{})
		q := "/?organization_name=acme-corp"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		app.newTeam(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	// TestTeam_GetHandler tests the getTeam handler. The getTeam page renders
	// permissions only if the authenticated user is an owner, so the test sets that
	// up first.
	t.Run("get", func(t *testing.T) {
		owners := newTestOwners(t, "acme-org")
		owner := NewUser(uuid.NewString(), WithTeams(owners))
		app := newFakeWeb(t, &fakeService{
			team:    owners,
			members: []*User{owner},
		})

		q := "/?team_id=team-123"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(otf.AddSubjectToContext(r.Context(), owner))
		w := httptest.NewRecorder()
		app.getTeam(w, r)
		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("update", func(t *testing.T) {
		team := NewTestTeam(t, "acme-org")
		app := newFakeWeb(t, &fakeService{
			team: team,
		})

		q := "/?team_id=team-123&manage_workspaces=true"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()
		app.updateTeam(w, r)
		html.AssertRedirect(t, w, paths.Team(team.ID))
	})

	t.Run("list", func(t *testing.T) {
		team := NewTestTeam(t, "acme-org")
		app := newFakeWeb(t, &fakeService{
			team: team,
		})

		q := "/?organization_name=acme"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(userCtx)
		w := httptest.NewRecorder()
		app.listTeams(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("delete", func(t *testing.T) {
		app := newFakeWeb(t, &fakeService{
			team: NewTestTeam(t, "acme-org"),
		})
		q := "/?team_id=team-123"
		r := httptest.NewRequest("POST", q, nil)
		w := httptest.NewRecorder()
		app.deleteTeam(w, r)
		html.AssertRedirect(t, w, paths.Teams("acme-org"))
	})
}

func TestUserDiff(t *testing.T) {
	a := []*User{{Username: "bob"}}
	b := []*User{{Username: "bob"}, {Username: "alice"}}
	assert.Equal(t, []*User{{Username: "alice"}}, diffUsers(a, b))
}
