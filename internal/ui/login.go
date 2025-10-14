package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/http/html"
)

type loginHandler struct {
	loginService loginService
}

type loginService interface {
	Clients() []*authenticator.OAuthClient
}

func addLoginHandlers(r *mux.Router, loginService loginService) {
	h := &loginHandler{
		loginService: loginService,
	}
	h.addHandlers(r)
}

func (h *loginHandler) addHandlers(r *mux.Router) {
	r.HandleFunc("/login", h.loginHandler)
}

func (h *loginHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(login(h.loginService.Clients()), w, r)
}
