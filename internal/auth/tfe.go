package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

const (
	addTeamMembersAction teamMembersAction = iota
	removeTeamMembersAction
)

type (
	teamMembersAction int

	tfe struct {
		AuthService
		*tfeapi.Responder
	}
)

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/account/details", a.getCurrentUser).Methods("GET")
	r.HandleFunc("/admin/users", a.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", a.deleteUser).Methods("DELETE")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.addTeamMembership).Methods("POST")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.removeTeamMembership).Methods("DELETE")

	r.HandleFunc("/organizations/{organization_name}/teams", a.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams", a.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", a.getTeamByName).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}", a.updateTeam).Methods("PATCH")
	r.HandleFunc("/teams/{team_id}", a.deleteTeam).Methods("DELETE")

	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/team-members
	r.HandleFunc("/teams/{team_id}/relationships/users", a.addTeamMembers).Methods("POST")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.removeTeamMembers).Methods("DELETE")

	// Stub implementation of the TFC Organization Memberships API:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organization-memberships
	//
	// Note: this is only implemented insofar as to satisfy some of the API tests we
	// run from the `go-tfe` project.
	r.HandleFunc("/organizations/{organization_name}/organization-memberships", a.inviteUser).Methods("POST")
	r.HandleFunc("/organization-memberships/{id}", a.deleteMembership).Methods("DELETE")
}

func (a *tfe) createUser(w http.ResponseWriter, r *http.Request) {
	var params types.CreateUserOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	user, err := a.CreateUser(r.Context(), *params.Username)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertUser(user), http.StatusCreated)
}

func (a *tfe) deleteUser(w http.ResponseWriter, r *http.Request) {
	username, err := decode.Param("username", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.DeleteUser(r.Context(), username); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) addTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   string `schema:"team_id,required"`
		Username string `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.AddTeamMembership(r.Context(), params.TeamID, []string{params.Username}); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *tfe) removeTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   string `schema:"team_id,required"`
		Username string `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.RemoveTeamMembership(r.Context(), params.TeamID, []string{params.Username}); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *tfe) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	a.Respond(w, r, a.convertUser(user), http.StatusOK)
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

	team, err := a.CreateTeam(r.Context(), org, opts)
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

	team, err := a.UpdateTeam(r.Context(), id, opts)
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

	teams, err := a.ListTeams(r.Context(), organization)
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

	team, err := a.GetTeam(r.Context(), *params.Organization, *params.Name)
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
	team, err := a.GetTeamByID(r.Context(), id)
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

	if err := a.DeleteTeam(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/team-members#add-a-user-to-team-with-user-id
func (a *tfe) addTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, addTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *tfe) removeTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, removeTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) modifyTeamMembers(r *http.Request, action teamMembersAction) error {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		return err
	}

	type teamMember struct {
		Username string `jsonapi:"primary,users"`
	}
	var users []teamMember
	if err := tfeapi.Unmarshal(r.Body, &users); err != nil {
		return err
	}

	// convert users into a simple slice of usernames
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	switch action {
	case addTeamMembersAction:
		return a.AddTeamMembership(r.Context(), teamID, usernames)
	case removeTeamMembersAction:
		return a.RemoveTeamMembership(r.Context(), teamID, usernames)
	default:
		return fmt.Errorf("unknown team membership action: %v", action)
	}
}

func (a *tfe) inviteUser(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.OrganizationMembershipCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	membership := &types.OrganizationMembership{
		ID: internal.NewID("ou"),
		User: &types.User{
			ID: internal.NewID("user"),
		},
		Organization: &types.Organization{
			Name: org,
		},
	}

	a.Respond(w, r, membership, http.StatusCreated)
}

func (a *tfe) deleteMembership(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convertUser(from *User) *types.User {
	return &types.User{
		ID:       from.ID,
		Username: from.Username,
	}
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

func (a *tfe) includeUsers(ctx context.Context, v any) ([]any, error) {
	team, ok := v.(*types.Team)
	if !ok {
		return nil, nil
	}
	users, err := a.ListTeamMembers(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	includes := make([]any, len(users))
	team.Users = make([]*types.User, len(users))
	for i, user := range users {
		team.Users[i] = &types.User{ID: user.ID}
		includes[i] = a.convertUser(user)
	}
	return includes, nil
}
