package agenttoken

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app appService
}

func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/agent/details", h.GetCurrentAgent).Methods("GET")
	r.HandleFunc("/agent/create", h.CreateAgentToken).Methods("POST")
}

func (h *handlers) CreateAgentToken(w http.ResponseWriter, r *http.Request) {
	opts := jsonapiCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	at, err := h.Application.CreateAgentToken(r.Context(), otf.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &AgentToken{at, true})
}

func (h *handlers) GetCurrentAgent(w http.ResponseWriter, r *http.Request) {
	agent, err := otf.AgentFromContext(r.Context())
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &AgentToken{agent, false})
}

type AgentToken struct {
	*otf.AgentToken
	revealToken bool // toggle send auth token over wire
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *AgentToken) ToJSONAPI() any {
	json := jsonapi.AgentToken{
		ID:           t.ID(),
		Organization: t.Organization(),
	}
	if t.revealToken {
		json.Token = t.Token()
	}
	return &json
}
