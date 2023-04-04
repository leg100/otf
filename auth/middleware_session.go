package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

const (
	// session cookie stores the session token
	sessionCookie = "session"
)

type AuthenticateSessionService interface {
	GetSession(ctx context.Context, token string) (*Session, error)
	GetUser(context.Context, UserSpec) (*User, error)
}

// AuthenticateSession verifies that all requests to /app endpoints possess
// a valid session cookie before attaching the corresponding user and session to
// the context.
func AuthenticateSession(svc AuthenticateSessionService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, paths.UIPrefix) {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(sessionCookie)
			if err == http.ErrNoCookie {
				html.SendUserToLoginPage(w, r)
				return
			}
			user, err := svc.GetUser(r.Context(), UserSpec{
				SessionToken: &cookie.Value,
			})
			if err != nil {
				html.FlashError(w, "unable to find user: "+err.Error())
				html.SendUserToLoginPage(w, r)
				return
			}

			session, err := svc.GetSession(r.Context(), cookie.Value)
			if err != nil {
				html.FlashError(w, "session expired")
				html.SendUserToLoginPage(w, r)
				return
			}

			// add user and session token to context for use by upstream handlers
			ctx := otf.AddSubjectToContext(r.Context(), user)
			ctx = addSessionCtx(ctx, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
