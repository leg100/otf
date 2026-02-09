package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/user"
)

type loginService interface {
	Clients() []*authenticator.OAuthClient
}

func addLoginHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/login", h.loginHandler)
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")
}

func (h *Handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(login(h.AuthenticatorService.Clients()), w, r)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *Handlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(h.templates.adminLogin(), w, r)
}

// adminLoginHandler logs in a site admin
func (h *Handlers) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if token != h.SiteToken {
		html.Error(r, w, "incorrect token")
		return
	}

	err = h.Tokens.StartSession(w, r, user.SiteAdminID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
}
