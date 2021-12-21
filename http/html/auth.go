package html

import (
	"context"
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2/github"
)

const (
	sessionUserKey  = "githubID"
	sessionUsername = "githubUsername"
	sessionFlashKey = "flash"
)

var (
	userSidebar = withSidebar("User Settings",
		sidebarItem{
			Name: "Profile",
			Link: "/profile",
		},
		sidebarItem{
			Name: "Sessions",
			Link: "/sessions",
		},
		sidebarItem{
			Name: "Tokens",
			Link: "/tokens",
		},
	)
)

type Profile struct {
	Username string
}

type Session struct {
	Token   string
	Expires time.Time
}

// githubLogin is called upon a successful Github login. A new user is created
// if they don't already exist.
func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
	githubUser, err := github.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessions.Put(r.Context(), sessionUserKey, *githubUser.ID)
	app.sessions.Put(r.Context(), sessionUsername, *githubUser.Login)

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (app *Application) isAuthenticated(r *http.Request) bool {
	return app.sessions.Exists(r.Context(), sessionUsername)
}

func (app *Application) requireAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.render(r, "login.tmpl", w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.sessions.Destroy(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessions.GetString(r.Context(), sessionUsername)
	prof := Profile{Username: username}

	if err := app.render(r, "profile.tmpl", w, &prof, userSidebar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	var sessions []Session

	currentUser := app.sessions.GetString(r.Context(), sessionUsername)

	err := app.sessions.Iterate(r.Context(), func(ctx context.Context) error {
		user := app.sessions.GetString(ctx, sessionUsername)
		if user == currentUser {
			sessions = append(sessions, Session{
				Token: app.sessions.Token(ctx),
			})
		}

		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.render(r, "sessions.tmpl", w, &sessions, userSidebar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
