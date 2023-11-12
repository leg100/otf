package team

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/tokens"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	svc           TeamService
	tokensService tokens.TokensService
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
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("team_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new team", org),
	})
}

func (h *webHandlers) createTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string
		Organization *string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.CreateTeam(r.Context(), *params.Organization, CreateTeamOptions{
		Name: params.Name,
	})
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "team already exists")
		http.Redirect(w, r, paths.NewTeam(*params.Organization), http.StatusFound)
		return
	}
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created team: "+team.Name)
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID string `schema:"team_id,required"`
		UpdateTeamOptions
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.UpdateTeam(r.Context(), params.TeamID, params.UpdateTeamOptions)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) listTeams(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := h.svc.ListTeams(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("team_list.tmpl", w, struct {
		organization.OrganizationPage
		Teams         []*Team
		CanCreateTeam bool
	}{
		OrganizationPage: organization.NewPage(r, "teams", org),
		Teams:            teams,
		CanCreateTeam:    subject.CanAccessOrganization(rbac.CreateTeamAction, org),
	})
}

func (h *webHandlers) deleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.GetTeamByID(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.svc.DeleteTeam(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted team: "+team.Name)
	http.Redirect(w, r, paths.Teams(team.Organization), http.StatusFound)
}
