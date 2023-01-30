package http

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (s *Server) CreateRegistrySession(w http.ResponseWriter, r *http.Request) {
	opts := jsonapi.RegistrySessionCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	session, err := s.Application.CreateRegistrySession(r.Context(), opts.OrganizationName)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &RegistrySession{session})
}

type RegistrySession struct {
	*otf.RegistrySession
}

// ToJSONAPI assembles a JSON-API DTO.
func (t *RegistrySession) ToJSONAPI() any {
	return &jsonapi.RegistrySession{
		Token:            t.Token(),
		OrganizationName: t.Organization(),
	}
}
