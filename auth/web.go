package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	svc AuthService
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	// Unauthenticated routes
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")

	// Authenticated routes
	r = html.UIRouter(r)

	h.addTeamHandlers(r)

	r.HandleFunc("/organizations/{name}/users", h.listUsers).Methods("GET")

	r.HandleFunc("/profile", h.profileHandler).Methods("GET")
}

func (h *webHandlers) listUsers(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := h.svc.ListUsers(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("users_list.tmpl", w, struct {
		html.SitePage
		Users []*User
	}{
		SitePage: html.NewSitePage(r, "users"),
		Users:    users,
	})
}

func (h *webHandlers) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("profile.tmpl", w, struct {
		html.SitePage
		User otf.Subject
	}{
		SitePage: html.NewSitePage(r, "profile"),
		User:     user,
	})
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *webHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("site_admin_login.tmpl", w, html.NewSitePage(r, "site admin login"))
}
