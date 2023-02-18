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

// web provides handlers for the web UI
type web struct {
	otf.Renderer

	app            application
	authenticators []*authenticator
	siteToken      string
}

func (h *web) addHandlers(r *mux.Router) {
	r.HandleFunc("/login", h.loginHandler)
	for _, auth := range h.authenticators {
		r.HandleFunc(auth.RequestPath(), auth.RequestHandler)
		r.HandleFunc(auth.CallbackPath(), auth.responseHandler)
	}

	//
	// Authenticated routes
	//
	r = r.NewRoute().Subrouter()
	r.Use(AuthenticateSession(h.app))

	h.addAgentTokenHandlers(r)
	h.addSessionHandlers(r)
	h.addTeamHandlers(r)

	r.HandleFunc("/organizations/{name}/users", h.listUsers).Methods("GET")

	r.HandleFunc("/logout", h.logoutHandler).Methods("POST")
	r.HandleFunc("/profile", h.profileHandler).Methods("POST")

	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLoginHandler).Methods("POST")
}

func (h *web) listUsers(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := h.app.listUsers(r.Context(), name)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("users_list.tmpl", w, r, users)
}

func (app *web) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}

func (app *web) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("login.tmpl", w, r, app.authenticators)
}

func (app *web) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSessionCtx(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.app.deleteSession(r.Context(), session.token); err != nil {
		return
	}
	html.SetCookie(w, sessionCookie, session.token, &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (app *web) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("site_admin_login.tmpl", w, r, nil)
}

// adminLoginHandler logs in a site admin
func (app *web) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != app.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	session, err := app.app.createSession(r, otf.SiteAdminID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set session cookie
	session.SetCookie(w)

	returnUserOriginalPage(w, r)
}
