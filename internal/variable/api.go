package variable

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	*tfeapi.Responder
	Client apiClient
}

type apiClient interface {
	ListEffectiveVariables(ctx context.Context, runID resource.TfeID) ([]*Variable, error)
}

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/vars/effective/{run_id}", a.listEffectiveVariables).Methods("GET")
}

func (a *API) listEffectiveVariables(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.Client.ListEffectiveVariables(r.Context(), runID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, variables, http.StatusOK)
}
