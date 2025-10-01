package team

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	teams      webClient
	tokens     *tokens.Service
	authorizer authz.Interface
}

type webClient interface {
	Create(ctx context.Context, organization organization.Name, opts CreateTeamOptions) (*Team, error)
	Get(ctx context.Context, organization organization.Name, team string) (*Team, error)
	GetByID(ctx context.Context, teamID resource.TfeID) (*Team, error)
	List(ctx context.Context, organization organization.Name) ([]*Team, error)
	Update(ctx context.Context, teamID resource.TfeID, opts UpdateTeamOptions) (*Team, error)
	Delete(ctx context.Context, teamID resource.TfeID) error
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/new", h.newTeam).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/create", h.createTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/delete", h.deleteTeam).Methods("POST")

	// NOTE: to avoid a import cycle the getTeam handler is instead located in
	// the user package.
}

func (h *webHandlers) newTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	html.Render(newTeamView(*params.Organization), w, r)
}

func (h *webHandlers) createTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.teams.Create(r.Context(), *params.Organization, CreateTeamOptions{
		Name: params.Name,
	})
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "team already exists")
		http.Redirect(w, r, paths.NewTeam(params.Organization), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created team: "+team.Name)
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID           resource.TfeID `schema:"team_id,required"`
		ManageWorkspaces bool           `schema:"manage_workspaces"`
		ManageVCS        bool           `schema:"manage_vcs"`
		ManageModules    bool           `schema:"manage_modules"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.teams.Update(r.Context(), params.TeamID, UpdateTeamOptions{
		OrganizationAccessOptions: OrganizationAccessOptions{
			ManageWorkspaces: &params.ManageWorkspaces,
			ManageVCS:        &params.ManageVCS,
			ManageModules:    &params.ManageModules,
		},
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) listTeams(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := h.teams.List(r.Context(), params.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := listTeamsProps{
		organization:  params.Organization,
		teams:         resource.NewPage(teams, params.PageOptions, nil),
		canCreateTeam: h.authorizer.CanAccess(r.Context(), authz.CreateTeamAction, params.Organization),
	}
	html.Render(listTeams(props), w, r)
}

func (h *webHandlers) deleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.teams.GetByID(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.teams.Delete(r.Context(), teamID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted team: "+team.Name)
	http.Redirect(w, r, paths.Teams(team.Organization), http.StatusFound)
}
