package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	API struct {
		*tfeapi.Responder
		Client apiClient
	}

	apiClient interface {
		Create(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
		Delete(ctx context.Context, username Username) error
		AddTeamMembership(ctx context.Context, teamID resource.ID, usernames []Username) error
		RemoveTeamMembership(ctx context.Context, teamID resource.ID, usernames []Username) error
	}

	modifyTeamMembershipOptions struct {
		Usernames []Username
	}
)

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/admin/users", a.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", a.deleteUser).Methods("DELETE")

	r.HandleFunc("/teams/{team_id}/relationships/users", a.addTeamMembers).Methods("POST")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.removeTeamMembers).Methods("DELETE")
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	var opts CreateUserOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	user, err := a.Client.Create(r.Context(), opts.Username)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, user, http.StatusCreated)
}

func (a *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Username Username `schema:"username"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.Delete(r.Context(), params.Username); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) addTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, addTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *API) removeTeamMembers(w http.ResponseWriter, r *http.Request) {
	if err := a.modifyTeamMembers(r, removeTeamMembersAction); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) modifyTeamMembers(r *http.Request, action teamMembersAction) error {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		return err
	}

	var opts modifyTeamMembershipOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		return err
	}

	switch action {
	case addTeamMembersAction:
		return a.Client.AddTeamMembership(r.Context(), teamID, opts.Usernames)
	case removeTeamMembersAction:
		return a.Client.RemoveTeamMembership(r.Context(), teamID, opts.Usernames)
	default:
		return fmt.Errorf("unknown team membership action: %v", action)
	}
}
