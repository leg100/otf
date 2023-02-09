package html

import (
	"net/http"

	"github.com/leg100/otf"
)

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.SubjectFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Render("profile.tmpl", w, r, user)
}
