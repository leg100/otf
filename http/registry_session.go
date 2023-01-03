package http

import (
	"net/http"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateRegistrySession(w http.ResponseWriter, r *http.Request) {
	opts := dto.RegistrySessionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	session, err := s.Application.CreateRegistrySession(r.Context(), opts.OrganizationName)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &RegistrySession{session})
}

type RegistrySession struct {
	*otf.RegistrySession
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *RegistrySession) ToJSONAPI() any {
	return &dto.RegistrySession{
		Token:            t.Token(),
		OrganizationName: t.Organization(),
	}
}
