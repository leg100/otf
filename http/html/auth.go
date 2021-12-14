package html

import (
	"net/http"

	"github.com/dghubble/gologin/v2/github"
)

const (
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
	sessionFlashKey = "flash"
)

type Profile struct {
	Username string
}

// issueSession issues a cookie session after successful Github login
func (app *application) issueSession() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		githubUser, err := github.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		app.sessions.Put(r.Context(), sessionUserKey, *githubUser.ID)
		app.sessions.Put(r.Context(), sessionUsername, *githubUser.Login)

		http.Redirect(w, r, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessions.Exists(r.Context(), sessionUsername)
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (app *application) profileHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessions.GetString(r.Context(), sessionUsername)
	prof := Profile{Username: username}

	if err := app.render(r, "profile.tmpl", w, &prof); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.render(r, "login.tmpl", w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.sessions.Destroy(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}
