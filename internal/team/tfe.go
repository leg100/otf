package team

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type TFEAPI struct {
	*tfeapi.Responder
	Client tfeClient
}

type tfeClient interface {
	CreateTeam(ctx context.Context, organization organization.Name, opts CreateTeamOptions) (*Team, error)
	GetTeam(ctx context.Context, organization organization.Name, name string) (*Team, error)
	GetTeamByID(ctx context.Context, teamID resource.TfeID) (*Team, error)
	ListTeams(ctx context.Context, organization organization.Name) ([]*Team, error)
	UpdateTeam(ctx context.Context, teamID resource.TfeID, opts UpdateTeamOptions) (*Team, error)
	DeleteTeam(ctx context.Context, teamID resource.TfeID) error

	CreateTeamToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error)
	GetTeamToken(ctx context.Context, teamID resource.TfeID) (*Token, error)
	DeleteTeamToken(ctx context.Context, teamID resource.TfeID) error
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
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

func (a *TFEAPI) createTeam(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFETeamCreateOptions
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

	team, err := a.Client.CreateTeam(r.Context(), pathParams.Organization, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusCreated)
}

func (a *TFEAPI) updateTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params TFETeamUpdateOptions
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

	team, err := a.Client.UpdateTeam(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *TFEAPI) listTeams(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	teams, err := a.Client.ListTeams(r.Context(), pathParams.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*TFETeam, len(teams))
	for i, from := range teams {
		items[i] = a.convertTeam(from)
	}
	a.Respond(w, r, items, http.StatusOK)
}

func (a *TFEAPI) getTeamByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *organization.Name `schema:"organization_name,required"`
		Name         *string            `schema:"team_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	team, err := a.Client.GetTeam(r.Context(), *params.Organization, *params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *TFEAPI) getTeam(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	team, err := a.Client.GetTeamByID(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertTeam(team), http.StatusOK)
}

func (a *TFEAPI) deleteTeam(w http.ResponseWriter, r *http.Request) {
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

func (a *TFEAPI) convertTeam(from *Team) *TFETeam {
	return &TFETeam{
		ID:         from.ID,
		Name:       from.Name,
		SSOTeamID:  from.SSOTeamID,
		Visibility: from.Visibility,
		OrganizationAccess: &TFEOrganizationAccess{
			ManageWorkspaces:      from.ManageWorkspaces,
			ManageVCSSettings:     from.ManageVCS,
			ManageModules:         from.ManageModules,
			ManageProviders:       from.ManageProviders,
			ManagePolicies:        from.ManagePolicies,
			ManagePolicyOverrides: from.ManagePolicyOverrides,
		},
		// Hardcode these values until proper support is added
		Permissions: &TFETeamPermissions{
			CanDestroy:          true,
			CanUpdateMembership: true,
		},
	}
}

func (a *TFEAPI) createTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.TeamTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	ot, token, err := a.Client.CreateTeamToken(r.Context(), CreateTokenOptions{
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

func (a *TFEAPI) getTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, err := a.Client.GetTeamToken(r.Context(), id)
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

func (a *TFEAPI) deleteTeamToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	err = a.Client.DeleteTeamToken(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
