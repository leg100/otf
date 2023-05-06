package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

func (a *api) addUserHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/account/details", a.getCurrentUser).Methods("GET")
	r.HandleFunc("/admin/users", a.createUser).Methods("POST")
	r.HandleFunc("/admin/users/{username}", a.deleteUser).Methods("DELETE")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.addTeamMembership).Methods("POST")
	r.HandleFunc("/teams/{team_id}/memberships/{username}", a.removeTeamMembership).Methods("DELETE")
}

func (a *api) createUser(w http.ResponseWriter, r *http.Request) {
	var params types.CreateUserOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	user, err := a.CreateUser(r.Context(), *params.Username)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, user, withCode(http.StatusCreated))
}

func (a *api) deleteUser(w http.ResponseWriter, r *http.Request) {
	username, err := decode.Param("username", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.DeleteUser(r.Context(), username); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) addTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params auth.TeamMembershipOptions
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}

	if err := a.AddTeamMembership(r.Context(), params); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *api) removeTeamMembership(w http.ResponseWriter, r *http.Request) {
	var params auth.TeamMembershipOptions
	if err := decode.Route(&params, r); err != nil {
		Error(w, err)
		return
	}

	if err := a.RemoveTeamMembership(r.Context(), params); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *api) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	a.writeResponse(w, r, user)
}
