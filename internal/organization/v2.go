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

	type v2org struct {
		Name string `json:"name"`
	}
	v2orgs := make([]v2org, len(orgs.Items))
	for i, org := range orgs.Items {
		v2orgs[i] = v2org{Name: org.Name}
	}

	if err := json.NewEncoder(w).Encode(v2orgs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
