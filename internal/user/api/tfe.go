package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	teampkg "github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/user"
)

type (
	TFEAPI struct {
		*tfeapi.Responder
		Client tfeClient
	}

	tfeClient interface {
		Create(ctx context.Context, username string, opts ...user.NewUserOption) (*user.User, error)
		ListOrganizationUsers(ctx context.Context, organization organization.Name) ([]*user.User, error)
		ListTeamUsers(ctx context.Context, teamID resource.TfeID) ([]*user.User, error)
		GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
		Delete(ctx context.Context, username user.Username) error
		AddTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error
		RemoveTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error
	}
)

func NewTFEAPI(client tfeClient, responder *tfeapi.Responder) *TFEAPI {
	api := &TFEAPI{
		Responder: responder,
		Client:    client,
	}

	// Fetch users when API calls request users be included in the
	// response
	responder.Register(tfeapi.IncludeUsers, api.includeUsers)

	return api
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
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

func (a *TFEAPI) addTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username user.Username  `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Client.AddTeamMembership(r.Context(), params.TeamID, []user.Username{params.Username}); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *TFEAPI) removeTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username user.Username  `schema:"username,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Client.RemoveTeamMembership(r.Context(), params.TeamID, []user.Username{params.Username}); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *TFEAPI) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := user.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	a.Respond(w, r, a.convertUser(user), http.StatusOK)
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/team-members#add-a-user-to-team-with-user-id
func (a *TFEAPI) addTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	type teamMember struct {
		Username user.Username `jsonapi:"primary,users"`
	}
	var users []teamMember
	if err := tfeapi.Unmarshal(r.Body, &users); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert users into a simple slice of usernames
	usernames := make([]user.Username, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	if err := a.Client.AddTeamMembership(r.Context(), teamID, usernames); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *TFEAPI) removeTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	type teamMember struct {
		Username user.Username `jsonapi:"primary,users"`
	}
	var users []teamMember
	if err := tfeapi.Unmarshal(r.Body, &users); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert users into a simple slice of usernames
	usernames := make([]user.Username, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	if err := a.Client.RemoveTeamMembership(r.Context(), teamID, usernames); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) inviteUser(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params user.TFEOrganizationMembershipCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	membership := &user.TFEOrganizationMembership{
		ID: resource.NewTfeID("ou"),
		User: &user.TFEUser{
			ID: resource.NewTfeID("user"),
		},
		Organization: &organization.TFEOrganization{
			Name: pathParams.Organization,
		},
	}

	a.Respond(w, r, membership, http.StatusCreated)
}

func (a *TFEAPI) deleteMembership(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convertUser(from *user.User) *user.TFEUser {
	return &user.TFEUser{
		ID:       from.ID,
		Username: from.Username.String(),
	}
}

func (a *TFEAPI) includeUsers(ctx context.Context, v any) ([]any, error) {
	team, ok := v.(*teampkg.TFETeam)
	if !ok {
		return nil, nil
	}
	users, err := a.Client.ListTeamUsers(ctx, team.ID)
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
