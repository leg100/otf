package html

import (
	"net/http"
	"path"

	"github.com/leg100/otf"
)

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

	// Get named user; if not exist create user
	user, err := app.UserService().Get(ctx, otf.UserSpecifier{Username: guser.Login})
	if err == otf.ErrResourceNotFound {
		user, err = app.UserService().Create(ctx, *guser.Login)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = app.sessions.TransferSession(ctx, user); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, app.route("getProfile"), http.StatusFound)
}

// requireAuthentication is middleware that insists on the user being
// authenticated before passing on the request.
func (app *Application) requireAuthentication(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !app.sessions.IsAuthenticated(r.Context()) {
			http.Redirect(w, r, app.route("login"), http.StatusFound)
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
	user := app.sessions.getUserFromContext(r.Context())

	tdata := app.newTemplateData(r, user)

	if err := app.renderTemplate("profile.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.sessions.getUserFromContext(r.Context())

	tdata := app.newTemplateData(r, user)

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

	if err := app.UserService().DeleteSession(r.Context(), token); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.sessions.FlashSuccess(r, "Revoked session"); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, app.route("listSession"), http.StatusFound)
}
