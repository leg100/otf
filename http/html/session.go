package html

import (
	"net/http"
	"sort"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
)

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, err := sessionFromContext(r.Context())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.ListSessions(r.Context(), user.ID())
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order sessions by creation date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt().After(sessions[j].CreatedAt())
	})

	app.Render("session_list.tmpl", w, r, struct {
		Items  []*otf.Session
		Active *otf.Session
	}{
		Items:  sessions,
		Active: active,
	})
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		Error(w, "missing token", http.StatusUnprocessableEntity)
		return
	}
	if err := app.DeleteSession(r.Context(), token); err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	FlashSuccess(w, "Revoked session")
	http.Redirect(w, r, paths.Sessions(), http.StatusFound)
}
