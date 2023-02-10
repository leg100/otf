package auth

import (
	"bytes"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/paths"
)

type htmlApp struct {
	otf.Renderer

	app app
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	// TODO: use session mw
	r.HandleFunc("/organizations/{organization_name}/users", app.listUsers).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/teams", app.listTeams).Methods("GET")
	r.HandleFunc("/teams/{team_id}", app.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", app.updateTeam).Methods("POST")

	// TODO: use session mw

	r.HandleFunc("/organizations/{organization_name}/agent-tokens", app.listAgentTokens)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/create", app.createAgentToken)
	r.HandleFunc("/organizations/{organization_name}/agent-tokens/new", app.newAgentToken)
	r.HandleFunc("/agent-tokens/{agent_token_id}/delete", app.deleteAgentToken)

	r.HandleFunc("/profile/sessions", app.sessionsHandler).Methods("GET")
	r.HandleFunc("/profile/sessions/revoke", app.revokeSessionHandler).Methods("POST")
	r.HandleFunc("/logout", app.logoutHandler).Methods("POST")
	r.HandleFunc("/profile", app.profileHandler).Methods("POST")

	// don't require authentication
	r.HandleFunc("/admin/login", app.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", app.adminLoginHandler).Methods("POST")
}

func (app *htmlApp) listUsers(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := app.app.listUsers(r.Context(), UserListOptions{
		Organization: otf.String(organization),
	})
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("users_list.tmpl", w, r, users)
}

func (app *htmlApp) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.app.getTeam(r.Context(), teamID)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := app.app.listTeamMembers(r.Context(), teamID)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("team_get.tmpl", w, r, struct {
		*Team
		Members []*User
	}{
		Team:    team,
		Members: members,
	})
}

func (app *htmlApp) updateTeam(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		TeamID string `schema:"team_id,required"`
		UpdateTeamOptions
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.app.updateTeam(r.Context(), params.TeamID, params.UpdateTeamOptions)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	otfhttp.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID()), http.StatusFound)
}

func (app *htmlApp) listTeams(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := app.app.listTeams(r.Context(), organization)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("team_list.tmpl", w, r, teams)
}

func (app *htmlApp) newAgentToken(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := app.GetOrganization(r.Context(), organization)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Render("agent_token_new.tmpl", w, r, org)
}

func (app *htmlApp) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts otf.CreateAgentTokenOptions
	if err := decode.All(&opts, r); err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	token, err := app.CreateAgentToken(r.Context(), opts)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// render a small templated flash message
	//
	// TODO: replace with a helper func, 'flashTemplate'
	buf := new(bytes.Buffer)
	if err := app.RenderTemplate("token_created.tmpl", buf, token.Token()); err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	otfhttp.FlashSuccess(w, buf.String())

	http.Redirect(w, r, paths.AgentTokens(opts.Organization), http.StatusFound)
}

func (app *htmlApp) listAgentTokens(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tokens, err := app.ListAgentTokens(r.Context(), organization)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
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

func (app *htmlApp) deleteAgentToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("agent_token_id", r)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	at, err := app.DeleteAgentToken(r.Context(), id)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	otfhttp.FlashSuccess(w, "Deleted token: "+at.Description())
	http.Redirect(w, r, paths.AgentTokens(at.Organization()), http.StatusFound)
}

func (app *htmlApp) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}

func (app *htmlApp) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, err := fromContext(r.Context())
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.app.list(r.Context(), user.ID())
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
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
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := app.app.delete(r.Context(), token); err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	otfhttp.FlashSuccess(w, "Revoked session")
	http.Redirect(w, r, paths.Sessions(), http.StatusFound)
}

func (app *htmlApp) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := fromContext(r.Context())
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.app.delete(r.Context(), session.Token()); err != nil {
		return
	}
	otfhttp.SetCookie(w, sessionCookie, session.Token(), &time.Time{})
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
		otfhttp.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != app.app.siteToken {
		otfhttp.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	session, err := app.app.CreateSession(r, otf.SiteAdminID)
	if err != nil {
		otfhttp.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set session cookie
	session.SetCookie(w)

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(otf.PathCookie); err == nil {
		otfhttp.SetCookie(w, otf.PathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}
