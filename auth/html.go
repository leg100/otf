package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type htmlApp struct {
	otf.Renderer

	app *Application
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	// TODO: use session mw
	r.HandleFunc("/organizations/{organization_name}/users", app.listUsers).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/teams", app.listTeams).Methods("GET")
	r.HandleFunc("/teams/{team_id}", app.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", app.updateTeam).Methods("POST")
}

func (app *htmlApp) listUsers(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := app.app.ListUsers(r.Context(), UserListOptions{
		Organization: otf.String(organization),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("users_list.tmpl", w, r, users)
}

func (app *htmlApp) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.app.GetTeam(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := app.app.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("team_get.tmpl", w, r, struct {
		*Team
		Members []*otf.User
	}{
		Team:    team,
		Members: members,
	})
}

func (app *htmlApp) updateTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts := UpdateTeamOptions{}
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.app.UpdateTeam(r.Context(), teamID, opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID()), http.StatusFound)
}

func (app *htmlApp) listTeams(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := app.app.ListTeams(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("team_list.tmpl", w, r, teams)
}
package agenttoken

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// webApp is the web application for organizations
type htmlHandlers struct {
	otf.Renderer // renders templates

	app appService // provide access to org service
}

func (app *htmlHandlers) AddHandlers(r *mux.Router) {
	// TODO: use session mw

	r.HandleFunc("/organizations/{organization_name}/agent-tokens", app.listAgentTokens)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", app.deleteAgentToken)
}

func (app *htmlHandlers) newAgentToken(w http.ResponseWriter, r *http.Request) {
	org, err := app.GetOrganization(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("agent_token_new.tmpl", w, r, org)
}

func (app *htmlHandlers) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	//
	// TODO: replace with a helper func, 'flashTemplate'
	buf := new(bytes.Buffer)
	if err := app.renderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (app *htmlHandlers) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.ListAgentTokens(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("agent_token_list.tmpl", w, r, struct {
		// list template expects pagination object but we don't paginate token
		// listing
		*otf.Pagination
		Items        []*otf.AgentToken
		Organization string
	}{
		Pagination:   &otf.Pagination{},
		Items:        tokens,
		Organization: organization,
	})
}

func (app *htmlHandlers) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := app.DeleteAgentToken(r.Context(), id)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "Deleted token: "+at.Description())
	http.Redirect(w, r, paths.AgentTokens(at.Organization()), http.StatusFound)
}
package session

import (
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type htmlApp struct {
	otf.Renderer

	app *Application
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/profile/sessions", app.sessionsHandler).Methods("GET")
	r.HandleFunc("/profile/sessions/revoke", app.revokeSessionHandler).Methods("POST")
	r.HandleFunc("/logout", app.logoutHandler).Methods("POST")
	r.HandleFunc("/profile", app.profileHandler).Methods("POST")

	// don't require authentication
	r.HandleFunc("/admin/login", app.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", app.adminLoginHandler).Methods("POST")
}

func (app *htmlApp) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}

func (app *htmlApp) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, err := fromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.app.list(r.Context(), user.ID())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order sessions by creation date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt().After(sessions[j].CreatedAt())
	})

	app.Render("session_list.tmpl", w, r, struct {
		Items  []*Session
		Active *Session
	}{
		Items:  sessions,
		Active: active,
	})
}

func (app *htmlApp) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := app.app.delete(r.Context(), token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Revoked session")
	http.Redirect(w, r, paths.Sessions(), http.StatusFound)
}

func (app *htmlApp) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := fromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.app.delete(r.Context(), session.Token()); err != nil {
		return
	}
	html.SetCookie(w, sessionCookie, session.Token(), &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (app *htmlApp) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("site_admin_login.tmpl", w, r, nil)
}

// adminLoginHandler logs in a site admin
func (app *htmlApp) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != app.app.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	session, err := app.app.CreateSession(r, otf.SiteAdminID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set session cookie
	session.SetCookie(w)

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(otf.PathCookie); err == nil {
		html.SetCookie(w, otf.PathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}
