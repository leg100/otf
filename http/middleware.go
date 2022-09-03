package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/leg100/otf"
)

// authTokenMiddleware checks the request has a valid api token
type authTokenMiddleware struct {
	otf.UserService
	otf.AgentTokenService
	siteToken string
}

func (m *authTokenMiddleware) handler(next http.Handler) http.Handler {
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

		user, err := m.isValid(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// add user to context for upstream handlers to consume
		ctx := otf.AddSubjectToContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *authTokenMiddleware) isValid(ctx context.Context, token string) (otf.Subject, error) {
	// check if site admin token
	if m.siteToken != "" {
		if m.siteToken == token {
			return &otf.SiteAdmin, nil
		}
	}

	// check if user token
	user, err := m.GetUser(ctx, otf.UserSpec{AuthenticationToken: &token})
	if err == nil {
		return user, nil
	}

	// check if agent token
	agentToken, err := m.GetAgentToken(ctx, token)
	if err == nil {
		return agentToken, nil
	}

	return nil, fmt.Errorf("invalid token")
}
