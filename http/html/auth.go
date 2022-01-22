package html

import (
	"net"
	"net/http"
	"path"

	"github.com/leg100/otf"
)

type Profile struct {
	Username string
}

// githubLogin is called upon a successful Github login. A new user is created
// if they don't already exist.
func (app *Application) githubLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token, err := app.oauth.responseHandler(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := app.oauth.newClient(ctx, token)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	guser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	anon := app.sessions.getUserFromContext(ctx)

	// promote anon user to auth user
	user, err := app.UserService().Promote(ctx, anon, *guser.Login)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
	app.sessions.Put(ctx, otf.AddressSessionKey, addr)

	http.Redirect(w, r, app.route("getProfile"), http.StatusFound)
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
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if err := app.sessions.Destroy(r.Context()); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (app *Application) meHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, path.Join(r.URL.Path, "profile"), http.StatusFound)
}

func (app *Application) profileHandler(w http.ResponseWriter, r *http.Request) {
	username := app.sessions.GetString(r.Context(), otf.UsernameSessionKey)

	tdata := app.newTemplateData(r, Profile{Username: username})

	if err := app.renderTemplate("profile.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.UserService().Get(r.Context(), app.currentUser(r))
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := app.newTemplateData(r, struct {
		ActiveToken string
		Sessions    []*otf.Session
	}{
		ActiveToken: app.sessions.Token(r.Context()),
		Sessions:    user.Sessions,
	})

	if err := app.renderTemplate("session_list.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) revokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		writeError(w, "missing token", http.StatusUnprocessableEntity)
		return
	}

	if err := app.UserService().RevokeSession(r.Context(), token, app.currentUser(r)); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessions.Put(r.Context(), otf.FlashSessionKey, "Revoked session")

	http.Redirect(w, r, "../", http.StatusFound)
}

func (app *Application) currentUser(r *http.Request) string {
	return app.sessions.GetString(r.Context(), otf.UsernameSessionKey)
}
