package html

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type currentOrganization struct {
	name string
}

// Name implements organizationName, for use in path helpers
func (c *currentOrganization) OrganizationName() string {
	return c.name
}

// authUser middleware ensures the request has a valid session cookie, attaching
// a session and user to the request context.
func (app *Application) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err == http.ErrNoCookie {
			http.Redirect(w, r, loginPath(), http.StatusFound)
			return
		}
		user, err := app.UserService().Get(r.Context(), otf.UserSpec{
			SessionToken: &cookie.Value,
		})
		if err != nil {
			flashError(w, "unable to find user: "+err.Error())
			http.Redirect(w, r, loginPath(), http.StatusFound)
			return
		}
		session := getActiveSession(user, cookie.Value)
		ctx := context.WithValue(r.Context(), userCtxKey, user)
		ctx = context.WithValue(ctx, sessionCtxKey, session)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// setOrganization ensures the session's organization reflects the most recently
// visited organization route.
func setOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current, ok := mux.Vars(r)["organization_name"]
		cookie, err := r.Cookie(organizationCookie)
		if ok {
			if err == http.ErrNoCookie || current != cookie.Value {
				// update session organization
				setCookie(w, organizationCookie, current, nil)
			}
		} else {
			if err == http.ErrNoCookie {
				// not yet visited an organization route
				next.ServeHTTP(w, r)
				return
			}
			// restore session org from cookie
			current = cookie.Value
		}
		// wrap session organization in context
		ctx := newOrganizationContext(r.Context(), current)
		r = r.WithContext(ctx)
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
