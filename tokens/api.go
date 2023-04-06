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

	// Registry session routes
	r.HandleFunc("/registry/sessions/create", h.createRegistrySession).Methods("POST")
}

// Registry session routes

func (h *api) createRegistrySession(w http.ResponseWriter, r *http.Request) {
	var opts jsonapi.RegistrySessionCreateOptions
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, err)
		return
	}

	token, err := h.svc.CreateRegistryToken(r.Context(), CreateRegistryTokenOptions{
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.Write(token)
}

// Agent token routes

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
