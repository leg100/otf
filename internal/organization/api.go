package organization

import (
	"net/http"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	Service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()

	r.HandleFunc("/api/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/api/organizations/{name}", a.deleteOrganization).Methods("DELETE")
}

func (a *api) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts CreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.CreateOrganization(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, org, http.StatusCreated)
}

func (a *api) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.DeleteOrganization(r.Context(), name); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
