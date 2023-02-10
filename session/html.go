package session

import (
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

type htmlApp struct {
	otf.Renderer

	app *Application
}

func (app *htmlApp) AddHTMLHandlers(r *mux.Router) {
	r.HandleFunc("/profile/sessions", app.sessionsHandler).Methods("GET")
	r.HandleFunc("/profile/sessions/revoke", app.revokeSessionHandler).Methods("POST")
	r.HandleFunc("/logout", app.logoutHandler).Methods("POST")
	r.HandleFunc("/profile", app.profileHandler).Methods("POST")

	// don't require authentication
	r.HandleFunc("/admin/login", app.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", app.adminLoginHandler).Methods("POST")
}

func (app *htmlApp) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}

func (app *htmlApp) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, err := fromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.app.list(r.Context(), user.ID())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order sessions by creation date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt().After(sessions[j].CreatedAt())
	})

	app.Render("session_list.tmpl", w, r, struct {
		Items  []*Session
		Active *Session
	}{
		Items:  sessions,
		Active: active,
	})
}

func (app *htmlApp) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := app.app.delete(r.Context(), token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Revoked session")
	http.Redirect(w, r, paths.Sessions(), http.StatusFound)
}

func (app *htmlApp) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := fromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.app.delete(r.Context(), session.Token()); err != nil {
		return
	}
	html.SetCookie(w, sessionCookie, session.Token(), &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (app *htmlApp) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("site_admin_login.tmpl", w, r, nil)
}

// adminLoginHandler logs in a site admin
func (app *htmlApp) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != app.app.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	session, err := app.app.CreateSession(r, otf.SiteAdminID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set session cookie
	session.SetCookie(w)

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(otf.PathCookie); err == nil {
		html.SetCookie(w, otf.PathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}
