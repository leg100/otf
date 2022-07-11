package http

import (
	"net/http"
)

func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeResponse(w, r, user)
}
