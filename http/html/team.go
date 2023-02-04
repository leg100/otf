package html

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.GetTeam(r.Context(), teamID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := app.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
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
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	opts := otf.UpdateTeamOptions{}
	if err := decode.All(&opts, r); err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := app.UpdateTeam(r.Context(), teamID, opts)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID()), http.StatusFound)
}

func (app *Application) listTeams(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := app.ListTeams(r.Context(), organization)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("team_list.tmpl", w, r, teams)
}
