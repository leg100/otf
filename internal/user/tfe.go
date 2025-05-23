package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	teampkg "github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
)

const (
	addTeamMembersAction teamMembersAction = iota
	removeTeamMembersAction
)

type (
	teamMembersAction int

	tfe struct {
		*Service
		*tfeapi.Responder
	}
)

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/account/details", a.getCurrentUser).Methods("GET")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.addTeamMembership).Methods("POST")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.removeTeamMembership).Methods("DELETE")

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

func (a *tfe) addTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username Username       `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.AddTeamMembership(r.Context(), params.TeamID, []Username{params.Username}); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *tfe) removeTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username Username       `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.RemoveTeamMembership(r.Context(), params.TeamID, []Username{params.Username}); err != nil {
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
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		return err
	}

	type teamMember struct {
		Username Username `jsonapi:"primary,users"`
	}
	var users []teamMember
	if err := tfeapi.Unmarshal(r.Body, &users); err != nil {
		return err
	}

	// convert users into a simple slice of usernames
	usernames := make([]Username, len(users))
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
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEOrganizationMembershipCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	membership := &TFEOrganizationMembership{
		ID: resource.NewTfeID("ou"),
		User: &TFEUser{
			ID: resource.NewTfeID("user"),
		},
		Organization: &organization.TFEOrganization{
			Name: pathParams.Organization,
		},
	}

	a.Respond(w, r, membership, http.StatusCreated)
}

func (a *tfe) deleteMembership(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convertUser(from *User) *TFEUser {
	return &TFEUser{
		ID:       from.ID,
		Username: from.Username.String(),
	}
}

func (a *tfe) includeUsers(ctx context.Context, v any) ([]any, error) {
	team, ok := v.(*teampkg.TFETeam)
	if !ok {
		return nil, nil
	}
	users, err := a.ListTeamUsers(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	includes := make([]any, len(users))
	team.Users = make([]*teampkg.TFEUser, len(users))
	for i, user := range users {
		team.Users[i] = &teampkg.TFEUser{ID: user.ID}
		includes[i] = a.convertUser(user)
	}
	return includes, nil
}
