package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/auth"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

const (
	addTeamMembersAction teamMembersAction = iota
	removeTeamMembersAction
)

type teamMembersAction int

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/team-members
func (a *api) addTeamMembershipHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/teams/{team_id}/relationships/users", a.addTeamMembers).Methods("POST")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.removeTeamMembers).Methods("DELETE")
}

// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/team-members#add-a-user-to-team-with-user-id
func (a *api) addTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, addTeamMembersAction); err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *api) removeTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, removeTeamMembersAction); err != nil {
		Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) modifyTeamMembers(r *http.Request, action teamMembersAction) error {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		return err
	}

	type teamMember struct {
		Username string `jsonapi:"primary,users"`
	}
	var users []teamMember
	if err := unmarshal(r.Body, &users); err != nil {
		return err
	}

	// convert users into a simple slice of usernames
	var usernames []string
	for _, u := range users {
		usernames = append(usernames, u.Username)
	}

	opts := auth.TeamMembershipOptions{
		TeamID:    teamID,
		Usernames: usernames,
	}
	switch action {
	case addTeamMembersAction:
		return a.AddTeamMembership(r.Context(), opts)
	case removeTeamMembersAction:
		return a.RemoveTeamMembership(r.Context(), opts)
	default:
		return fmt.Errorf("unknown team membership action: %v", action)
	}
}
