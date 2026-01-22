package variable

import (
	"net/http"

	otfhttp "github.com/leg100/otf/internal/http"

	"github.com/leg100/otf/internal/tfeapi"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

type api struct {
	*Service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()
	r.HandleFunc("/vars/effective/{run_id}", a.listEffectiveVariables).Methods("GET")
}

func (a *api) listEffectiveVariables(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.ID("run_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.ListEffectiveVariables(r.Context(), runID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, variables, http.StatusOK)
}
