package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	api struct {
		*Service
		*tfeapi.Responder
	}

	modifyTeamMembershipOptions struct {
		Usernames []Username
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()

	r.HandleFunc("/admin/users", a.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", a.deleteUser).Methods("DELETE")

	r.HandleFunc("/teams/{team_id}/relationships/users", a.addTeamMembers).Methods("POST")
	r.HandleFunc("/teams/{team_id}/relationships/users", a.removeTeamMembers).Methods("DELETE")
}

func (a *api) createUser(w http.ResponseWriter, r *http.Request) {
	var opts CreateUserOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	user, err := a.Create(r.Context(), opts.Username)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, user, http.StatusCreated)
}

func (a *api) deleteUser(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Username Username `schema:"username"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Delete(r.Context(), params.Username); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
		return a.AddTeamMembership(r.Context(), teamID, opts.Usernames)
	case removeTeamMembersAction:
		return a.RemoveTeamMembership(r.Context(), teamID, opts.Usernames)
	default:
		return fmt.Errorf("unknown team membership action: %v", action)
	}
}
