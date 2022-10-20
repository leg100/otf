package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) getTeam(w http.ResponseWriter, r *http.Request) {
	spec := otf.TeamSpec{
		OrganizationName: mux.Vars(r)["organization_name"],
		Name:             mux.Vars(r)["team_name"],
	}
	team, err := app.GetTeam(r.Context(), spec.Name, spec.OrganizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := app.ListUsers(r.Context(), otf.UserListOptions{
		OrganizationName: otf.String(spec.OrganizationName),
		TeamName:         otf.String(spec.Name),
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("team_get.tmpl", w, r, struct {
		*otf.Team
		Members []*otf.User
	}{
		Team:    team,
		Members: members,
	})
}

func (app *Application) updateTeam(w http.ResponseWriter, r *http.Request) {
	spec := otf.TeamSpec{
		OrganizationName: mux.Vars(r)["organization_name"],
		Name:             mux.Vars(r)["team_name"],
	}
	opts := otf.TeamUpdateOptions{}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
	}
	team, err := app.UpdateTeam(r.Context(), spec.Name, spec.OrganizationName, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "team permissions updated")
	http.Redirect(w, r, getTeamPath(team), http.StatusFound)
}

func (app *Application) listTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := app.ListTeams(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("team_list.tmpl", w, r, teams)
}

func (app *Application) listTeamUsers(w http.ResponseWriter, r *http.Request) {
	opts := otf.UserListOptions{
		OrganizationName: otf.String(mux.Vars(r)["organization_name"]),
		TeamName:         otf.String(mux.Vars(r)["team_name"]),
	}
	users, err := app.ListUsers(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("team_users_list.tmpl", w, r, UserList{users, opts})
}
