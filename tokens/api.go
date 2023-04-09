package tokens

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
)

// api provides handlers for json:api endpoints
type api struct {
	svc TokensService
}

func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	// Agent token routes
	r.HandleFunc("/agent/details", h.getCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.createAgentToken).Methods("POST")

	// Run token routes
	r.HandleFunc("/tokens/run/create", h.createRunToken).Methods("POST")
}

func (h *api) createRunToken(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.CreateRunTokenOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}

	token, err := h.svc.CreateRunToken(r.Context(), CreateRunTokenOptions{
		Organization: opts.Organization,
		RunID:        opts.RunID,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.Write(token)
}

func (h *api) createAgentToken(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.AgentTokenCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}
	token, err := h.svc.CreateAgentToken(r.Context(), CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	w.Write(token)
}

func (h *api) getCurrentAgent(w http.ResponseWriter, r *http.Request) {
	at, err := agentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	jsonapi.WriteResponse(w, r, &jsonapi.AgentToken{
		ID:           at.ID,
		Organization: at.Organization,
	})
}
