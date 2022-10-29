package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// authUser middleware ensures the request has a valid session cookie, attaching
// a session and user to the request context.
type authMiddleware struct {
	users    otf.UserService
	sessions otf.SessionService
}

func (m *authMiddleware) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err == http.ErrNoCookie {
			sendUserToLoginPage(w, r)
			return
		}
		user, err := m.users.GetUser(r.Context(), otf.UserSpec{
			SessionToken: &cookie.Value,
		})
		if err != nil {
			flashError(w, "unable to find user: "+err.Error())
			sendUserToLoginPage(w, r)
			return
		}
		session, err := m.sessions.GetSessionByToken(r.Context(), cookie.Value)
		if err != nil {
			flashError(w, "unable to find session: "+err.Error())
			sendUserToLoginPage(w, r)
			return
		}
		ctx := otf.AddSubjectToContext(r.Context(), user)
		ctx = addSessionToContext(ctx, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	setCookie(w, pathCookie, r.URL.Path, nil)
	http.Redirect(w, r, loginPath(), http.StatusFound)
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
