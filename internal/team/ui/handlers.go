package ui

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
)

type Handlers struct {
	Client     Client
	Authorizer authz.Interface
}

type Client interface {
	CreateTeam(ctx context.Context, organization organization.Name, opts team.CreateTeamOptions) (*team.Team, error)
	GetTeamByID(ctx context.Context, teamID resource.ID) (*team.Team, error)
	ListTeams(ctx context.Context, organization organization.Name) ([]*team.Team, error)
	UpdateTeam(ctx context.Context, teamID resource.ID, opts team.UpdateTeamOptions) (*team.Team, error)
	DeleteTeam(ctx context.Context, teamID resource.ID) error
	List(ctx context.Context) ([]*user.User, error)
	ListTeamUsers(ctx context.Context, teamID resource.ID) ([]*user.User, error)
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/new", h.newTeam).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/create", h.createTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}", h.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/delete", h.deleteTeam).Methods("POST")
}

func (h *Handlers) newTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	helpers.RenderPage(
		newTeamView(*params.Organization),
		"new team",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Teams", Link: paths.Teams(params.Organization)},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) createTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string
		Organization *organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	createdTeam, err := h.Client.CreateTeam(r.Context(), *params.Organization, team.CreateTeamOptions{
		Name: params.Name,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created team: "+createdTeam.Name)
	http.Redirect(w, r, paths.Team(createdTeam.ID), http.StatusFound)
}

func (h *Handlers) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID           resource.TfeID `schema:"team_id,required"`
		ManageWorkspaces bool           `schema:"manage_workspaces"`
		ManageVCS        bool           `schema:"manage_vcs"`
		ManageModules    bool           `schema:"manage_modules"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	updatedTeam, err := h.Client.UpdateTeam(r.Context(), params.TeamID, team.UpdateTeamOptions{
		OrganizationAccessOptions: team.OrganizationAccessOptions{
			ManageWorkspaces: &params.ManageWorkspaces,
			ManageVCS:        &params.ManageVCS,
			ManageModules:    &params.ManageModules,
		},
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(updatedTeam.ID), http.StatusFound)
}

func (h *Handlers) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	team, err := h.Client.GetTeamByID(r.Context(), teamID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// get usernames of team members
	members, err := h.Client.ListTeamUsers(r.Context(), teamID)
	if err != nil {
		helpers.Error(r, w, err.Error())
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
	if h.Authorizer.CanAccess(r.Context(), authz.ListUsersAction, resource.SiteID) {
		users, err := h.Client.List(r.Context())
		if err != nil {
			helpers.Error(r, w, err.Error())
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
		canUpdateTeam:   h.Authorizer.CanAccess(r.Context(), authz.UpdateTeamAction, team.Organization),
		canDeleteTeam:   h.Authorizer.CanAccess(r.Context(), authz.DeleteTeamAction, team.Organization),
		canAddMember:    h.Authorizer.CanAccess(r.Context(), authz.AddTeamMembershipAction, team.Organization),
		canRemoveMember: h.Authorizer.CanAccess(r.Context(), authz.RemoveTeamMembershipAction, team.Organization),
		dropdown: helpers.SearchDropdownProps{
			Name:        "username",
			Available:   internal.ConvertSliceToString(nonMemberUsernames),
			Existing:    internal.ConvertSliceToString(usernames),
			Action:      templ.SafeURL(paths.AddMemberTeam(team.ID)),
			Placeholder: "Add user",
			Width:       helpers.WideDropDown,
		},
	}
	helpers.RenderPage(
		getTeam(props),
		team.ID.String(),
		w,
		r,
		helpers.WithOrganization(team.Organization),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Teams", Link: paths.Teams(props.team.Organization)},
			helpers.Breadcrumb{Name: props.team.Name},
		),
	)
}

func (h *Handlers) listTeams(w http.ResponseWriter, r *http.Request) {
	var params team.ListOptions
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	teams, err := h.Client.ListTeams(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := listTeamsProps{
		organization:  params.Organization,
		teams:         resource.NewPage(teams, params.PageOptions, nil),
		canCreateTeam: h.Authorizer.CanAccess(r.Context(), authz.CreateTeamAction, params.Organization),
	}
	helpers.RenderPage(
		listTeams(props),
		"teams",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithContentActions(listTeamsActions(props)),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Teams"}),
	)
}

func (h *Handlers) deleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	deletedTeam, err := h.Client.GetTeamByID(r.Context(), teamID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	err = h.Client.DeleteTeam(r.Context(), teamID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "deleted team: "+deletedTeam.Name)
	http.Redirect(w, r, paths.Teams(deletedTeam.Organization), http.StatusFound)
}

// diffUsers returns the users from b that are not in a.
func diffUsers(a, b []*user.User) (c []*user.User) {
	m := make(map[user.Username]struct{}, len(a))
	for _, u := range a {
		m[u.Username] = struct{}{}
	}
	for _, u := range b {
		if _, ok := m[u.Username]; !ok {
			c = append(c, u)
		}
	}
	return
}
