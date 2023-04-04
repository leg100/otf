package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"google.golang.org/api/idtoken"
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

	authTokenMiddleware struct {
		AuthenticateTokenService
		AuthenticateTokenConfig
	}
)

// AuthenticateToken verifies that all requests to /api/v2 endpoints possess
// a valid bearer token.
func NewAuthTokenMiddleware(svc AuthenticateTokenService, cfg AuthenticateTokenConfig) mux.MiddlewareFunc {
	mw := authTokenMiddleware{
		AuthenticateTokenService: svc,
		AuthenticateTokenConfig:  cfg,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !requiresToken(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			subj, err := mw.isValid(r)
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

func (m *authTokenMiddleware) isValid(r *http.Request) (otf.Subject, error) {
	ctx := r.Context()

	if auth := r.Header.Get("Authorization"); auth != "" {
		splitToken := strings.Split(auth, "Bearer ")
		if len(splitToken) != 2 {
			return nil, fmt.Errorf("malformed bearer token")
		}
		token := splitToken[1]

		switch {
		case strings.HasPrefix(token, "agent."):
			return m.GetAgentToken(ctx, token)
		case strings.HasPrefix(token, "registry."):
			return m.GetRegistrySession(ctx, token)
		case strings.HasPrefix(token, "user."):
			return m.GetUser(ctx, UserSpec{AuthenticationToken: &token})
		case m.SiteToken != "" && m.SiteToken == token:
			return &SiteAdmin, nil
		case m.GoogleJWTConfig.Enabled && m.GoogleJWTConfig.Header == "":
			return m.validateJWT(ctx, token)
		default:
			return nil, fmt.Errorf("invalid bearer token")
		}
	} else if m.GoogleJWTConfig.Enabled && m.Header != "" {
		return m.validateJWT(ctx, r.Header.Get(m.Header))
	} else {
		return nil, fmt.Errorf("no authentication token found")
	}
}

func (m *authTokenMiddleware) validateJWT(ctx context.Context, token string) (otf.Subject, error) {
	payload, err := idtoken.Validate(ctx, token, m.Audience)
	if err != nil {
		return nil, err
	}
	return m.GetUser(ctx, UserSpec{Username: &payload.Subject})
}

func requiresToken(path string) bool {
	for _, prefix := range otfhttp.AuthenticatedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
