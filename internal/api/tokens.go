package api

import (
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/tokens"
)

func (a *api) addTokenHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Agent token routes
	r.HandleFunc("/agent/details", a.getCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", a.createAgentToken).Methods("POST")

	// Run token routes
	r.HandleFunc("/tokens/run/create", a.createRunToken).Methods("POST")
}

func (a *api) createRunToken(w http.ResponseWriter, r *http.Request) {
	var opts types.CreateRunTokenOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	token, err := a.CreateRunToken(r.Context(), tokens.CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		Error(w, err)
		return
	}

	w.Write(token)
}

func (a *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts types.AgentTokenCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	token, err := a.CreateAgentToken(r.Context(), tokens.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Write(token)
}

func (a *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := tokens.AgentFromContext(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	b, err := jsonapi.Marshal(&types.AgentToken{
		ID:           at.ID,
		Organization: at.Organization,
	})
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.Write(b)
}
