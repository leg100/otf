package agent

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type api struct {
	*service
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfapi.DefaultBasePath).Subrouter()
	r.HandleFunc("/agent/register", a.registerAgent).Methods("POST")
}

func (a *api) registerAgent(w http.ResponseWriter, r *http.Request) {
	// middleware should have put agent token into context.
	token, err := tokens.AgentFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	var params struct {
		Name        *string // optional name
		Concurrency int
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	agent, err := a.service.registerAgent(r.Context(), registerAgentOptions{
		Name:        params.Name,
		Concurrency: params.Concurrency,
		IPAddress:   net.ParseIP(r.RemoteAddr),
		AgentPoolID: &token.AgentPoolID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.Respond(w, r, agent, http.StatusCreated)
}
