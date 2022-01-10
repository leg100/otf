package html

import (
	"net"
	"net/http"

	"github.com/leg100/otf"
)

var (
	userSidebar = withSidebar("user settings",
		anchor{
			Name: "profile",
			Link: "/profile",
		},
		anchor{
			Name: "sessions",
			Link: "/sessions",
		},
		anchor{
			Name: "tokens",
			Link: "/tokens",
		},
	)
)

type Profile struct {
	Username string
}

// githubLogin is called upon a successful Github login. A new user is created
// if they don't already exist.
func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
	token, err := app.oauth.responseHandler(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := app.oauth.newClient(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, _, err := client.Users.Get(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We cannot rely on the LoadAndSave() middleware to save session token to
	// DB because it only does so after this handler has finished, but Login()
	// below relies on it having already been saved so we do so now.
	_, _, err = app.sessions.Commit(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	opts := otf.UserLoginOptions{
		Username:     *user.Login,
		SessionToken: app.sessions.Token(r.Context()),
	}

	if err := app.UserService().Login(r.Context(), opts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Populate session data
	app.sessions.Put(r.Context(), otf.UsernameSessionKey, *user.Login)

	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	app.sessions.Put(r.Context(), otf.AddressSessionKey, addr)

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (app *Application) isAuthenticated(r *http.Request) bool {
	return app.sessions.Exists(r.Context(), otf.UsernameSessionKey)
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
	tdata := app.newTemplateData(r, nil)

	if err := app.renderTemplate("login.tmpl", w, tdata); err != nil {
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
	username := app.sessions.GetString(r.Context(), otf.UsernameSessionKey)

	tdata := app.newTemplateData(r, Profile{Username: username})

	if err := app.renderTemplate("profile.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.UserService().Get(r.Context(), app.currentUser(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := app.newTemplateData(r, struct {
		ActiveToken string
		Sessions    []*otf.Session
	}{
		ActiveToken: app.sessions.Token(r.Context()),
		Sessions:    user.Sessions,
	})

	if err := app.renderTemplate("sessions.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnprocessableEntity)
		return
	}

	if err := app.UserService().RevokeSession(r.Context(), token, app.currentUser(r)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessions.Put(r.Context(), otf.FlashSessionKey, "Revoked session")

	http.Redirect(w, r, "/sessions", http.StatusFound)
}

func (app *Application) currentUser(r *http.Request) string {
	return app.sessions.GetString(r.Context(), otf.UsernameSessionKey)
}
