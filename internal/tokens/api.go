package tokens

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	TokensService
	*tfeapi.Responder
}

func (a *api) addHandlers(r *mux.Router) {
	r.HandleFunc("/agent/create", a.createAgentToken).Methods("POST")
	r.HandleFunc("/agent/details", a.getCurrentAgent).Methods("GET")
	r.HandleFunc("/tokens/run/create", a.createRunToken).Methods("POST")
}

func (a *api) createRunToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateRunTokenOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.CreateRunToken(r.Context(), CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateAgentTokenOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	token, err := a.CreateAgentToken(r.Context(), opts)
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
