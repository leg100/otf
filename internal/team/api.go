package team

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	API struct {
		*tfeapi.Responder
		Client apiClient
	}

	apiClient interface {
		CreateTeam(ctx context.Context, organization organization.Name, opts CreateTeamOptions) (*Team, error)
		GetTeam(ctx context.Context, organization organization.Name, name string) (*Team, error)
		DeleteTeam(ctx context.Context, teamID resource.ID) error
	}
)

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/teams", a.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", a.getTeamByName).Methods("GET")

	r.HandleFunc("/teams/{team_id}", a.deleteTeam).Methods("DELETE")
}

func (a *API) createTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	var opts CreateTeamOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	team, err := a.Client.CreateTeam(r.Context(), params.Name, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, team, http.StatusCreated)
}

func (a *API) getTeamByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		Team         string            `schema:"team_name,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	team, err := a.Client.GetTeam(r.Context(), params.Organization, params.Team)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, team, http.StatusOK)
}

func (a *API) deleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Client.DeleteTeam(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
