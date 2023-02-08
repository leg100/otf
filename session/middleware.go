package session

import (
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

const (
	// session cookie stores the session identifier
	sessionCookie = "session"
)

// authUser middleware ensures the request has a valid session cookie, attaching
// a session and user to the request context.
type authMiddleware struct {
	otf.Application
}

func (m *authMiddleware) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err == http.ErrNoCookie {
			sendUserToLoginPage(w, r)
			return
		}
		user, err := m.GetUser(r.Context(), otf.UserSpec{
			SessionToken: &cookie.Value,
		})
		if err != nil {
			html.FlashError(w, "unable to find user: "+err.Error())
			sendUserToLoginPage(w, r)
			return
		}
		session, err := m.GetSessionByToken(r.Context(), cookie.Value)
		if err != nil {
			html.FlashError(w, "unable to find session: "+err.Error())
			sendUserToLoginPage(w, r)
			return
		}
		ctx := otf.AddSubjectToContext(r.Context(), user)
		ctx = addToContext(ctx, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, otf.PathCookie, r.URL.Path, nil)
	http.Redirect(w, r, paths.Login(), http.StatusFound)
}
