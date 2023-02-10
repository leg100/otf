package auth

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
package agenttoken

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

func TestAgentToken_NewHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		org: org,
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newAgentToken(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_CreateHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	token := otf.NewTestAgentToken(t, org)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		token: token,
	})

	q := "/?organization_name=acme&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.createAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens("acme"), redirect.Path)
	}
}

func TestAgentToken_ListHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	token := otf.NewTestAgentToken(t, org)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		org:   org,
		token: token,
	})

	q := "/?organization_name=acme"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listAgentTokens(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeAgentTokenHandlerApp{
		token: otf.NewTestAgentToken(t, org),
	})

	q := "/?agent_token_id=at-123"
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()
	app.deleteAgentToken(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, paths.AgentTokens(org.Name()), redirect.Path)
	}
}

type fakeAgentTokenHandlerApp struct {
	org   *otf.Organization
	token *otf.AgentToken

	otf.Application
}

func (f *fakeAgentTokenHandlerApp) GetOrganization(context.Context, string) (*otf.Organization, error) {
	return f.org, nil
}

func (f *fakeAgentTokenHandlerApp) CreateAgentToken(context.Context, otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return f.token, nil
}

func (f *fakeAgentTokenHandlerApp) ListAgentTokens(context.Context, string) ([]*otf.AgentToken, error) {
	return []*otf.AgentToken{f.token}, nil
}

func (f *fakeAgentTokenHandlerApp) DeleteAgentToken(context.Context, string) (*otf.AgentToken, error) {
	return f.token, nil
}
package session

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionHandlers(t *testing.T) {
	user := otf.NewTestUser(t)
	active := otf.NewTestSession(t, user.ID())
	other := otf.NewTestSession(t, user.ID())

	app := newFakeWebApp(t, &fakeSessionHandlerApp{sessions: []*otf.Session{active, other}})

	t.Run("list sessions", func(t *testing.T) {
		// add user and active session to request
		r := httptest.NewRequest("GET", "/sessions", nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		r = r.WithContext(addToContext(r.Context(), active))

		w := httptest.NewRecorder()
		app.sessionsHandler(w, r)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("revoke session", func(t *testing.T) {
		form := strings.NewReader(url.Values{
			"token": {"asklfdkljfj"},
		}.Encode())

		r := httptest.NewRequest("POST", "/sessions/delete", form)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		app.revokeSessionHandler(w, r)

		assert.Equal(t, 302, w.Code)
	})
}

func TestAdminLoginHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeAdminLoginHandlerApp{}, withFakeSiteToken("secrettoken"))

	tests := []struct {
		name         string
		token        string
		wantRedirect string
	}{
		{
			name:         "valid token",
			token:        "secrettoken",
			wantRedirect: "/profile",
		},
		{
			name:         "invalid token",
			token:        "badtoken",
			wantRedirect: "/admin/login",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := strings.NewReader(url.Values{
				"token": {tt.token},
			}.Encode())

			r := httptest.NewRequest("POST", "/admin/login", form)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			app.adminLoginHandler(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, tt.wantRedirect, redirect.Path)
			}
		})
	}
}

type fakeSessionHandlerApp struct {
	sessions []*otf.Session
	otf.Application
}

func (f *fakeSessionHandlerApp) ListSessions(context.Context, string) ([]*otf.Session, error) {
	return f.sessions, nil
}

func (f *fakeSessionHandlerApp) DeleteSession(context.Context, string) error {
	return nil
}
