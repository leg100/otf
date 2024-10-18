package organization

import (
	"encoding/json"
	"net/http"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/gorilla/mux"
)

type v2 struct {
	*Service
}

func (a *v2) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.V2BasePath).Subrouter()

	r.HandleFunc("/organizations", a.listOrganizations).Methods("GET")
}

func (a *v2) listOrganizations(w http.ResponseWriter, r *http.Request) {
	// TODO: process pagination parameters
	orgs, err := a.List(r.Context(), ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(orgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
