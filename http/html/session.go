package html

import (
	"net/http"
	"sort"

	"github.com/leg100/otf"
)

// sessionList exposes a list of sessions to a template
type sessionList struct {
	// list template expects pagination object but we don't paginate session
	// listing
	*otf.Pagination
	Items []*otf.Session
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// display sessions by creation date, newest first
	sort.Slice(user.Sessions, func(i, j int) bool {
		return user.Sessions[i].CreatedAt().After(user.Sessions[j].CreatedAt())
	})
	app.render("session_list.tmpl", w, r, sessionList{
		Pagination: &otf.Pagination{},
		Items:      user.Sessions,
	})
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		writeError(w, "missing token", http.StatusUnprocessableEntity)
		return
	}
	if err := app.UserService().DeleteSession(r.Context(), token); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Revoked session")
	http.Redirect(w, r, listSessionPath(), http.StatusFound)
}
