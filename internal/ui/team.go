package ui

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
)

type (
	teamHandlers struct {
		teams      teamClient
		tokens     *tokens.Service
		authorizer teamAuthorizer
	}

	teamClient interface {
		Create(ctx context.Context, organization organization.Name, opts team.CreateTeamOptions) (*team.Team, error)
		Get(ctx context.Context, organization organization.Name, teamName string) (*team.Team, error)
		GetByID(ctx context.Context, teamID resource.TfeID) (*team.Team, error)
		List(ctx context.Context, organization organization.Name) ([]*team.Team, error)
		Update(ctx context.Context, teamID resource.TfeID, opts team.UpdateTeamOptions) (*team.Team, error)
		Delete(ctx context.Context, teamID resource.TfeID) error
	}

	teamAuthorizer interface {
		CanAccess(context.Context, authz.Action, resource.ID) bool
	}
)

// addTeamHandlers registers team UI handlers with the router
func addTeamHandlers(r *mux.Router, teams teamClient, tokens *tokens.Service, authorizer teamAuthorizer) {
	h := &teamHandlers{
		authorizer: authorizer,
		teams:      teams,
		tokens:     tokens,
	}

	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/new", h.newTeam).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/create", h.createTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/delete", h.deleteTeam).Methods("POST")

	// NOTE: to avoid a import cycle the getTeam handler is instead located in
	// the user package.
}

func (h *teamHandlers) newTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	html.Render(newTeamView(*params.Organization), w, r)
}

func (h *teamHandlers) createTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	createdTeam, err := h.teams.Create(r.Context(), *params.Organization, team.CreateTeamOptions{
		Name: params.Name,
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created team: "+createdTeam.Name)
	http.Redirect(w, r, paths.Team(createdTeam.ID), http.StatusFound)
}

func (h *teamHandlers) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID           resource.TfeID `schema:"team_id,required"`
		ManageWorkspaces bool           `schema:"manage_workspaces"`
		ManageVCS        bool           `schema:"manage_vcs"`
		ManageModules    bool           `schema:"manage_modules"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	updatedTeam, err := h.teams.Update(r.Context(), params.TeamID, team.UpdateTeamOptions{
		OrganizationAccessOptions: team.OrganizationAccessOptions{
			ManageWorkspaces: &params.ManageWorkspaces,
			ManageVCS:        &params.ManageVCS,
			ManageModules:    &params.ManageModules,
		},
	})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(updatedTeam.ID), http.StatusFound)
}

func (h *teamHandlers) listTeams(w http.ResponseWriter, r *http.Request) {
	var params team.ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	teams, err := h.teams.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	props := listTeamsProps{
		organization:  params.Organization,
		teams:         resource.NewPage(teams, params.PageOptions, nil),
		canCreateTeam: h.authorizer.CanAccess(r.Context(), authz.CreateTeamAction, params.Organization),
	}
	html.Render(listTeams(props), w, r)
}

func (h *teamHandlers) deleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	deletedTeam, err := h.teams.GetByID(r.Context(), teamID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	err = h.teams.Delete(r.Context(), teamID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "deleted team: "+deletedTeam.Name)
	http.Redirect(w, r, paths.Teams(deletedTeam.Organization), http.StatusFound)
}
