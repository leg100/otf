package auth

import (
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

func (h *web) addSessionHandlers(r *mux.Router) {
	r.HandleFunc("/profile/sessions", h.sessionsHandler).Methods("GET")
	r.HandleFunc("/profile/sessions/revoke", h.revokeSessionHandler).Methods("POST")
}

func (app *web) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := otf.UserFromContext(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, err := getSessionCtx(r.Context())
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessions, err := app.svc.listSessions(r.Context(), user.ID())
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

func (app *web) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := app.svc.deleteSession(r.Context(), token); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Revoked session")
	http.Redirect(w, r, paths.Sessions(), http.StatusFound)
}
