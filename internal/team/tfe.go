package team

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/organizations/{organization_name}/teams", a.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams", a.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", a.getTeamByName).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.updateTeam).Methods("PATCH")
	r.HandleFunc("/teams/{team_id}", a.deleteTeam).Methods("DELETE")

	// Team token routes
	r.HandleFunc("/teams/{team_id}/authentication-token", a.createTeamToken).Methods("POST")
	r.HandleFunc("/teams/{team_id}/authentication-token", a.getTeamToken).Methods("GET")
	r.HandleFunc("/teams/{team_id}/authentication-token", a.deleteTeamToken).Methods("DELETE")

}

func (a *tfe) createTeam(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params types.TeamCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := CreateTeamOptions{
		Name:       params.Name,
		SSOTeamID:  params.SSOTeamID,
		Visibility: params.Visibility,
	}
	if params.OrganizationAccess != nil {
		opts.OrganizationAccessOptions = OrganizationAccessOptions{
			ManageWorkspaces:      params.OrganizationAccess.ManageWorkspaces,
			ManageVCS:             params.OrganizationAccess.ManageVCSSettings,
			ManageModules:         params.OrganizationAccess.ManageModules,
			ManageProviders:       params.OrganizationAccess.ManageProviders,
			ManagePolicies:        params.OrganizationAccess.ManagePolicies,
			ManagePolicyOverrides: params.OrganizationAccess.ManagePolicyOverrides,
		}
	}

	team, err := a.Create(r.Context(), org, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusCreated)
}

func (a *tfe) updateTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params types.TeamUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := UpdateTeamOptions{
		Name:       params.Name,
		SSOTeamID:  params.SSOTeamID,
		Visibility: params.Visibility,
	}
	if params.OrganizationAccess != nil {
		opts.OrganizationAccessOptions = OrganizationAccessOptions{
			ManageWorkspaces:      params.OrganizationAccess.ManageWorkspaces,
			ManageVCS:             params.OrganizationAccess.ManageVCSSettings,
			ManageModules:         params.OrganizationAccess.ManageModules,
			ManageProviders:       params.OrganizationAccess.ManageProviders,
			ManagePolicies:        params.OrganizationAccess.ManagePolicies,
			ManagePolicyOverrides: params.OrganizationAccess.ManagePolicyOverrides,
		}
	}

	team, err := a.Update(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *tfe) listTeams(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	teams, err := a.List(r.Context(), organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*types.Team, len(teams))
	for i, from := range teams {
		items[i] = a.convertTeam(from)
	}
	a.Respond(w, r, items, http.StatusOK)
}

func (a *tfe) getTeamByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *string `schema:"organization_name,required"`
		Name         *string `schema:"team_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	team, err := a.Get(r.Context(), *params.Organization, *params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *tfe) getTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	team, err := a.GetByID(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *tfe) deleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Delete(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convertTeam(from *Team) *types.Team {
	return &types.Team{
		ID:         from.ID,
		Name:       from.Name,
		SSOTeamID:  from.SSOTeamID,
		Visibility: from.Visibility,
		OrganizationAccess: &types.OrganizationAccess{
			ManageWorkspaces:      from.Access.ManageWorkspaces,
			ManageVCSSettings:     from.Access.ManageVCS,
			ManageModules:         from.Access.ManageModules,
			ManageProviders:       from.Access.ManageProviders,
			ManagePolicies:        from.Access.ManagePolicies,
			ManagePolicyOverrides: from.Access.ManagePolicyOverrides,
		},
		// Hardcode these values until proper support is added
		Permissions: &types.TeamPermissions{
			CanDestroy:          true,
			CanUpdateMembership: true,
		},
	}
}

func (a *tfe) createTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.TeamTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	ot, token, err := a.CreateTeamToken(r.Context(), CreateTokenOptions{
		TeamID: id,
		Expiry: opts.ExpiredAt,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := &types.TeamToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		Token:     string(token),
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) getTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, err := a.GetTeamToken(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if ot == nil {
		tfeapi.Error(w, internal.ErrResourceNotFound)
		return
	}

	to := &types.TeamToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) deleteTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	err = a.DeleteTeamToken(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
