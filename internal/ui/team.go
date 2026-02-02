package ui

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/user"
)

type (
	teamHandlers struct {
		teams      teamClient
		users      usersClient
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
func addTeamHandlers(r *mux.Router, teams teamClient, users usersClient, tokens *tokens.Service, authorizer teamAuthorizer) {
	h := &teamHandlers{
		authorizer: authorizer,
		teams:      teams,
		users:      users,
		tokens:     tokens,
	}

	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/new", h.newTeam).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/create", h.createTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}", h.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/delete", h.deleteTeam).Methods("POST")
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

func (h *teamHandlers) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	team, err := h.teams.GetByID(r.Context(), teamID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	// get usernames of team members
	members, err := h.users.ListTeamUsers(r.Context(), teamID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
	usernames := make([]user.Username, len(members))
	for i, m := range members {
		usernames[i] = m.Username
	}

	// Retrieve full list of users for populating a select form from which new
	// team members can be chosen. Only do this if the subject has perms to
	// retrieve the list.
	var nonMemberUsernames []user.Username
	if h.authorizer.CanAccess(r.Context(), authz.ListUsersAction, resource.SiteID) {
		users, err := h.users.List(r.Context())
		if err != nil {
			html.Error(r, w, err.Error())
			return
		}
		nonMembers := diffUsers(members, users)
		nonMemberUsernames = make([]user.Username, len(nonMembers))
		for i, m := range nonMembers {
			nonMemberUsernames[i] = m.Username
		}
	}

	props := getTeamProps{
		team:            team,
		members:         members,
		canUpdateTeam:   h.authorizer.CanAccess(r.Context(), authz.UpdateTeamAction, team.Organization),
		canDeleteTeam:   h.authorizer.CanAccess(r.Context(), authz.DeleteTeamAction, team.Organization),
		canAddMember:    h.authorizer.CanAccess(r.Context(), authz.AddTeamMembershipAction, team.Organization),
		canRemoveMember: h.authorizer.CanAccess(r.Context(), authz.RemoveTeamMembershipAction, team.Organization),
		dropdown: helpers.SearchDropdownProps{
			Name:        "username",
			Available:   internal.ConvertSliceToString(nonMemberUsernames),
			Existing:    internal.ConvertSliceToString(usernames),
			Action:      templ.SafeURL(paths.AddMemberTeam(team.ID)),
			Placeholder: "Add user",
			Width:       helpers.WideDropDown,
		},
	}
	html.Render(getTeam(props), w, r)
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
