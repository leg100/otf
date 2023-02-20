package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

func (h *web) addTeamHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/teams/{team_id}", h.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
}

func (h *web) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.app.getTeamByID(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := h.app.listTeamMembers(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("team_get.tmpl", w, r, struct {
		*Team
		Members []*User
	}{
		Team:    team,
		Members: members,
	})
}

func (h *web) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID string `schema:"team_id,required"`
		UpdateTeamOptions
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.app.updateTeam(r.Context(), params.TeamID, params.UpdateTeamOptions)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID()), http.StatusFound)
}

func (h *web) listTeams(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := h.app.listTeams(r.Context(), organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("team_list.tmpl", w, r, teams)
}
