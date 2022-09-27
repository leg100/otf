package http

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeResponse(w, r, &User{user})
}

type User struct {
	*otf.User
}

// ToJSONAPI assembles a JSON-API DTO.
func (u *User) ToJSONAPI() any {
	return &dto.User{
		ID:       u.ID(),
		Username: u.Username(),
	}
}
