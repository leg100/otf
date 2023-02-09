package user

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
	r.HandleFunc("/account/details", h.GetCurrentUser).Methods("GET")
}

func (h *handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonapi.WriteResponse(w, r, &User{user})
}
