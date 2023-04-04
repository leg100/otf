package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"google.golang.org/api/idtoken"
)

const (
	// session cookie stores the session token
	sessionCookie = "session"
)

type (
	AuthenticateTokenService interface {
		GetAgentToken(context.Context, string) (*AgentToken, error)
		GetRegistrySession(context.Context, string) (*RegistrySession, error)
		GetUser(ctx context.Context, spec UserSpec) (*User, error)
	}

	AuthenticateTokenConfig struct {
		SiteToken string
		GoogleJWTConfig
	}

	GoogleJWTConfig struct {
		Enabled  bool
		Header   string
		Audience string
	}
)

// AuthenticateToken verifies that all requests to /api/v2 endpoints possess
// a valid bearer token.
func AuthenticateToken(svc AuthenticateTokenService, cfg AuthenticateTokenConfig) mux.MiddlewareFunc {
	isValid := func(r *http.Request, token string) (otf.Subject, error) {
		ctx := r.Context()
		switch {
		case strings.HasPrefix(token, "agent."):
			return svc.GetAgentToken(ctx, token)
		case strings.HasPrefix(token, "registry."):
			return svc.GetRegistrySession(ctx, token)
		case strings.HasPrefix(token, "user."):
			return svc.GetUser(ctx, UserSpec{AuthenticationToken: &token})
		case cfg.SiteToken != "" && cfg.SiteToken == token:
			return &SiteAdmin, nil
		case cfg.GoogleJWTConfig.Enabled:
			if cfg.Header != "" {
				token = r.Header.Get(cfg.Header)
			}
			payload, err := idtoken.Validate(ctx, token, cfg.Audience)
			if err != nil {
				return nil, err
			}
			return svc.GetUser(ctx, UserSpec{Username: &payload.Subject})
		default:
			return nil, fmt.Errorf("no auth token found")
		}
	}

	requiresToken := func(path string) bool {
		for _, prefix := range otfhttp.AuthenticatedPrefixes {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requiresToken(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
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

			subj, err := isValid(r, token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// add subject to context for upstream handlers to consume
			ctx := otf.AddSubjectToContext(r.Context(), subj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
