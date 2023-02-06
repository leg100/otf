package user

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &User{user})
}
