package organization

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	*tfeapi.Responder
	Client apiClient
}

type apiClient interface {
	CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error)
	DeleteOrganization(ctx context.Context, name Name) error
}

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/organizations/{name}", a.deleteOrganization).Methods("DELETE")
}

func (a *API) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts CreateOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	org, err := a.Client.CreateOrganization(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, org, http.StatusCreated)
}

func (a *API) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.DeleteOrganization(r.Context(), params.Name); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
