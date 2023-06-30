package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

func (a *api) addTeamHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/teams", a.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", a.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.getTeamByID).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.updateTeam).Methods("PATCH")
	r.HandleFunc("/teams/{team_id}", a.deleteTeam).Methods("DELETE")
}

func (a *api) createTeam(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	var params types.TeamCreateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	team, err := a.CreateTeam(r.Context(), auth.CreateTeamOptions{
		Name:         *params.Name,
		Organization: org,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, team, withCode(http.StatusCreated))
}

func (a *api) updateTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	var params types.TeamUpdateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	team, err := a.UpdateTeam(r.Context(), id, auth.UpdateTeamOptions{
		Name: params.Name,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, team)
}

func (a *api) getTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *string `schema:"organization_name,required"`
		Name         *string `schema:"team_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	team, err := a.GetTeam(r.Context(), *params.Organization, *params.Name)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, team)
}

func (a *api) getTeamByID(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	team, err := a.GetTeamByID(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, team)
}

func (a *api) deleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.DeleteTeam(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
