package http

import (
	"net/http"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateAgentToken(w http.ResponseWriter, r *http.Request) {
	opts := dto.AgentTokenCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	at, err := s.Application.CreateAgentToken(r.Context(), otf.CreateAgentTokenOptions{
		Description:  opts.Description,
		Organization: opts.Organization,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &AgentToken{at, true})
}

func (s *Server) GetCurrentAgent(w http.ResponseWriter, r *http.Request) {
	agent, err := otf.AgentFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &AgentToken{agent, false})
}

type AgentToken struct {
	*otf.AgentToken
	revealToken bool // toggle send auth token over wire
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *AgentToken) ToJSONAPI() any {
	json := dto.AgentToken{
		ID:           t.ID(),
		Organization: t.Organization(),
	}
	if t.revealToken {
		json.Token = t.Token()
	}
	return &json
}
