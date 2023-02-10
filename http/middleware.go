package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
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
