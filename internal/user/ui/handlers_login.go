package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/user"
)

type LoginHandlers struct {
	Client    loginClient
	SiteToken string
}

type loginClient interface {
	Clients() []*authenticator.OAuthClient
	StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error
}

func (h *LoginHandlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/login", h.loginHandler)
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")
}

func (h *LoginHandlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	helpers.Render(login(h.Client.Clients()), w, r)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *LoginHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	helpers.Render(adminLogin(), w, r)
}

// adminLoginHandler logs in a site admin
func (h *LoginHandlers) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	if token != h.SiteToken {
		helpers.Error(r, w, "incorrect token")
		return
	}

	err = h.Client.StartSession(w, r, user.SiteAdminID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
}
