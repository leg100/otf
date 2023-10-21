package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	api struct {
		AuthService
		*tfeapi.Responder
	}

	modifyTeamMembershipOptions struct {
		Usernames []string
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()

	r.HandleFunc("/admin/users", a.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", a.deleteUser).Methods("DELETE")

	r.HandleFunc("/organizations/{organization_name}/teams", a.createTeam).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/teams/{team_name}", a.getTeamByName).Methods("GET")

	r.HandleFunc("/teams/{team_id}", a.deleteTeam).Methods("DELETE")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.addTeamMembers).Methods("POST")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.removeTeamMembers).Methods("DELETE")
}

func (a *api) createUser(w http.ResponseWriter, r *http.Request) {
	var opts CreateUserOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	user, err := a.CreateUser(r.Context(), opts.Username)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, user, http.StatusCreated)
}

func (a *api) deleteUser(w http.ResponseWriter, r *http.Request) {
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

func (a *api) createTeam(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var opts CreateTeamOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	team, err := a.CreateTeam(r.Context(), org, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, team, http.StatusCreated)
}

func (a *api) getTeamByName(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization string `schema:"organization_name,required"`
		Team         string `schema:"team_name,required"`
	}
	if err := decode.Route(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	team, err := a.GetTeam(r.Context(), params.Organization, params.Team)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, team, http.StatusOK)
}

func (a *api) deleteTeam(w http.ResponseWriter, r *http.Request) {
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
func (a *api) addTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, addTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *api) removeTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, removeTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) modifyTeamMembers(r *http.Request, action teamMembersAction) error {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		return err
	}

	var opts modifyTeamMembershipOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		return err
	}

	switch action {
	case addTeamMembersAction:
		return a.AddTeamMembership(r.Context(), teamID, opts.Usernames)
	case removeTeamMembersAction:
		return a.RemoveTeamMembership(r.Context(), teamID, opts.Usernames)
	default:
		return fmt.Errorf("unknown team membership action: %v", action)
	}
}
