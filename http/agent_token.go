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
		Organization: opts.OrganizationName,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &AgentToken{at})
}

func (s *Server) GetCurrentAgent(w http.ResponseWriter, r *http.Request) {
	agent, err := otf.AgentFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	agent.HideToken()
	writeResponse(w, r, &AgentToken{agent})
}

type AgentToken struct {
	*otf.AgentToken
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *AgentToken) ToJSONAPI() any {
	return &dto.AgentToken{
		ID:               t.ID(),
		Token:            t.Token(),
		OrganizationName: t.Organization(),
	}
}
