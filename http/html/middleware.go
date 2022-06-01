package html

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// authUser middleware ensures the request has a valid session cookie, attaching
// a session and user to the request context.
func (app *Application) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err == http.ErrNoCookie {
			http.Redirect(w, r, app.route("login"), http.StatusFound)
			return
		}
		user, err := app.UserService().Get(r.Context(), otf.UserSpec{
			SessionToken: &cookie.Value,
		})
		if err != nil {
			flashError(w, "unable to find user: "+err.Error())
			http.Redirect(w, r, app.route("login"), http.StatusFound)
			return
		}
		session := getActiveSession(user, cookie.Value)
		ctx := context.WithValue(r.Context(), userCtxKey, user)
		ctx = context.WithValue(ctx, sessionCtxKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// setCurrentOrganization ensures a user's current organization matches the
// organization in the request. If there is no organization in the current
// request then no action is taken.
func (app *Application) setCurrentOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := getCtxUser(r.Context())
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		current, ok := mux.Vars(r)["organization_name"]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		if user.CurrentOrganization == nil || *user.CurrentOrganization != current {
			user.CurrentOrganization = &current
			if err := app.UserService().SetCurrentOrganization(r.Context(), user.ID(), current); err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func getActiveSession(user *otf.User, token string) *otf.Session {
	for _, session := range user.Sessions {
		if session.Token == token {
			return session
		}
	}
	panic("no active session found")
}
