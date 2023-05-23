package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
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

	r.HandleFunc("/organizations/{name}/users", h.listOrganizationUsers).Methods("GET")

	r.HandleFunc("/profile", h.profileHandler).Methods("GET")
}

func (h *webHandlers) listOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := h.svc.ListOrganizationUsers(r.Context(), name)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
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
	user, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("profile.tmpl", w, struct {
		html.SitePage
		User internal.Subject
	}{
		SitePage: html.NewSitePage(r, "profile"),
		User:     user,
	})
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *webHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("site_admin_login.tmpl", w, html.NewSitePage(r, "site admin login"))
}
