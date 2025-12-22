package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
)

type loginHandlers struct {
	loginService loginService
	tokens       tokensClient
	siteToken    string
}

type loginService interface {
	Clients() []*authenticator.OAuthClient
}

type tokensClient interface {
	StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error
}

func addLoginHandlers(r *mux.Router, loginService loginService, tokens tokensClient, siteToken string) {
	h := &loginHandlers{
		loginService: loginService,
		tokens:       tokens,
		siteToken:    siteToken,
	}
	r.HandleFunc("/login", h.loginHandler)
	r.HandleFunc("/admin/login", adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")
}

func (h *loginHandlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(login(h.loginService.Clients()), w, r)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	html.Render(adminLogin(), w, r)
}

// adminLoginHandler logs in a site admin
func (h *loginHandlers) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if token != h.siteToken {
		html.Error(r, w, "incorrect token")
		return
	}

	err = h.tokens.StartSession(w, r, user.SiteAdminID)
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}
}
