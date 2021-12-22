package html

import (
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2/github"
	"github.com/leg100/otf"
)

const (
	sessionUsername = "username"
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
	guser, err := github.UserFromContext(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opts := otf.UserLoginOptions{
		Username:     *guser.Login,
		SessionToken: app.sessions.Token(r.Context()),
	}

	if err := app.UserService().Login(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	app.sessions.Put(r.Context(), sessionUsername, *guser.Login)

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
	user, err := app.UserService().Get(r.Context(), app.currentUser(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.render(r, "sessions.tmpl", w, &user.Sessions, userSidebar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) currentUser(r *http.Request) string {
	return app.sessions.GetString(r.Context(), sessionUsername)
}
