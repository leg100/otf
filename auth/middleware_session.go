package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	// session cookie stores the session token
	sessionCookie = "session"
)

type AuthenticateSessionService interface {
	GetUser(context.Context, UserSpec) (*User, error)
}

// NewAuthSessionMiddleware constructs middleware that verifies all requests to
// /app endpoints possess a valid session cookie before attaching the
// corresponding to the context.
func NewAuthSessionMiddleware(svc AuthenticateSessionService, secret string) (mux.MiddlewareFunc, error) {
	key, err := jwk.FromRaw([]byte(secret))
	if err != nil {
		return nil, err
	}
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
			// parse jwt from cookie and verify signature
			token, err := jwt.Parse([]byte(cookie.Value), jwt.WithKey(jwa.HS256, key))
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired()) {
					html.FlashError(w, "session expired")
				} else {
					html.FlashError(w, "unable to verify session token: "+err.Error())
				}
				html.SendUserToLoginPage(w, r)
				return
			}
			user, err := svc.GetUser(r.Context(), UserSpec{
				Username: otf.String(token.Subject()),
			})
			if err != nil {
				html.FlashError(w, "unable to find user: "+err.Error())
				html.SendUserToLoginPage(w, r)
				return
			}
			// add user and session token to context for use by upstream handlers
			ctx := otf.AddSubjectToContext(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
