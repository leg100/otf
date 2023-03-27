package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	otf.Renderer

	svc            AuthService
	authenticators []*authenticator
	siteToken      string
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	// Unauthenticated routes
	r.HandleFunc("/login", h.loginHandler)
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")

	for _, auth := range h.authenticators {
		r.HandleFunc(auth.RequestPath(), auth.RequestHandler)
		r.HandleFunc(auth.CallbackPath(), auth.responseHandler)
	}

	// Authenticated routes
	r = html.UIRouter(r)
	r.Use(AuthenticateSession(h.svc)) // require session cookie

	h.addAgentTokenHandlers(r)
	h.addSessionHandlers(r)
	h.addTeamHandlers(r)
	h.addTokenHandlers(r)

	r.HandleFunc("/organizations/{name}/users", h.listUsers).Methods("GET")

	r.HandleFunc("/logout", h.logoutHandler).Methods("POST")
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

	h.Render("users_list.tmpl", w, r, users)
}

func (h *webHandlers) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("profile.tmpl", w, r, user)
}

func (h *webHandlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("login.tmpl", w, r, h.authenticators)
}

func (h *webHandlers) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSessionCtx(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.svc.DeleteSession(r.Context(), session.token); err != nil {
		return
	}
	html.SetCookie(w, sessionCookie, session.token, &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *webHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("site_admin_login.tmpl", w, r, nil)
}

// adminLoginHandler logs in a site admin
func (h *webHandlers) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != h.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	session, err := h.svc.CreateSession(r.Context(), CreateSessionOptions{
		Request: r,
		UserID:  otf.String(otf.SiteAdminID),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.setCookie(w)

	returnUserOriginalPage(w, r)
}
