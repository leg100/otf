package organization

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	*Service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()

	r.HandleFunc("/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/organizations/{name}", a.deleteOrganization).Methods("DELETE")
}

func (a *api) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts CreateOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	org, err := a.Create(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, org, http.StatusCreated)
}

func (a *api) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Delete(r.Context(), params.Name); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
