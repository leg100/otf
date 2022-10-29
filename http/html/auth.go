package html

import (
	"net"
	"net/http"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.render("login.tmpl", w, r, app.authenticators)
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := app.DeleteSession(r.Context(), session.Token()); err != nil {
		return
	}
	setCookie(w, sessionCookie, session.Token(), &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("profile.tmpl", w, r, user)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (app *Application) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	app.render("site_admin_login.tmpl", w, r, nil)
}

// adminLoginHandler logs in a site admin
func (app *Application) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	// expect token in POST form
	type adminLoginForm struct {
		Token *string `schema:"token,required"`
	}
	var form adminLoginForm
	if err := decode.Form(&form, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if *form.Token != app.siteToken {
		flashError(w, "incorrect token")
		http.Redirect(w, r, adminLoginPath(), http.StatusFound)
		return
	}

	if err := createSession(app, w, r, otf.SiteAdminID); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return admin to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		setCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, getProfilePath(), http.StatusFound)
	}
}

// createSession creates a session for the user with the given user ID
func createSession(app otf.Application, w http.ResponseWriter, r *http.Request, uid string) error {
	// get admin's source IP address
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err
	}

	session, err := app.CreateSession(r.Context(), uid, addr)
	if err != nil {
		return err
	}
	setCookie(w, sessionCookie, session.Token(), otf.Time(session.Expiry()))
	return nil
}
