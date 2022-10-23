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
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.ListSessions(r.Context(), user.ID())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order sessions by creation date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt().After(sessions[j].CreatedAt())
	})

	app.render("session_list.tmpl", w, r, sessionList{
		Pagination: &otf.Pagination{},
		Items:      sessions,
	})
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		writeError(w, "missing token", http.StatusUnprocessableEntity)
		return
	}
	if err := app.DeleteSession(r.Context(), token); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashSuccess(w, "Revoked session")
	http.Redirect(w, r, listSessionPath(), http.StatusFound)
}
