package html

import (
	"net/http"
	"time"

	"github.com/leg100/otf"
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
	if err := app.DeleteSession(r.Context(), session.Token); err != nil {
		return
	}
	setCookie(w, sessionCookie, session.Token, &time.Time{})
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
