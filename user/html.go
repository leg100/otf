package user

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
