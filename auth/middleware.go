package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/paths"
)

const (
	// session cookie stores the session token
	sessionCookie = "session"
)

type AuthenticateTokenService interface {
	GetUser(context.Context, otf.UserSpec) (otf.User, error)
	GetAgentToken(context.Context, string) (otf.AgentToken, error)
	GetRegistrySession(context.Context, string) (otf.RegistrySession, error)
}

// AuthenticateToken checks the request has a valid API token
func AuthenticateToken(svc AuthenticateTokenService) mux.MiddlewareFunc {
	isValid := func(ctx context.Context, token string) (otf.Subject, error) {
		switch {
		case strings.HasPrefix(token, "agent."):
			return svc.GetAgentToken(ctx, token)
		case strings.HasPrefix(token, "registry."):
			return svc.GetRegistrySession(ctx, token)
		default:
			// otherwise assume user or site admin token
			return svc.GetUser(ctx, otf.UserSpec{AuthenticationToken: &token})
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHdr := r.Header.Get("Authorization")
			if authHdr == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			hdr := strings.Split(authHdr, "Bearer ")
			if len(hdr) != 2 {
				http.Error(w, "malformed token", http.StatusUnauthorized)
				return
			}
			token := hdr[1]

			subj, err := isValid(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// add subject to context for upstream handlers to consume
			ctx := otf.AddSubjectToContext(r.Context(), subj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticateSession middleware checks incoming request possesses a valid session cookie,
// attaching its user and the session to the context.
func AuthenticateSession(svc otf.UserService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err == http.ErrNoCookie {
				sendUserToLoginPage(w, r)
				return
			}
			user, err := svc.GetUser(r.Context(), otf.UserSpec{
				SessionToken: &cookie.Value,
			})
			if err != nil {
				otfhttp.FlashError(w, "unable to find user: "+err.Error())
				sendUserToLoginPage(w, r)
				return
			}

			// add user and session token to context for use by upstream handlers
			ctx := otf.AddSubjectToContext(r.Context(), user)
			ctx = addToContext(ctx, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	otfhttp.SetCookie(w, otf.PathCookie, r.URL.Path, nil)
	http.Redirect(w, r, paths.Login(), http.StatusFound)
}
