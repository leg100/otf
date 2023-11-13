package remoteops

import (
	"encoding/json"
	"net/http"

	otfapi "github.com/leg100/otf/internal/api"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	svc AgentTokenService
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()
	r.HandleFunc("/agent/create", a.createAgentToken).Methods("POST")
	r.HandleFunc("/agent/details", a.getCurrentAgent).Methods("GET")
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentTokenOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.svc.CreateAgentToken(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := AgentFromContext(r.Context())
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, at, http.StatusOK)
}
