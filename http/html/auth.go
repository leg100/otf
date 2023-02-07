package html

import (
	"net/http"

	"github.com/leg100/otf"
)

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("login.tmpl", w, r, app.authenticators)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (app *Application) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	app.Render("site_admin_login.tmpl", w, r, nil)
}
