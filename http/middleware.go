package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/leg100/otf"
)

// authToken checks the request has a valid api token
type authTokenMiddleware struct {
	svc       otf.UserService
	siteToken string
}

func (m *authTokenMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(hdr) != 2 {
			http.Error(w, "malformed token", http.StatusUnprocessableEntity)
			return
		}
		token := hdr[1]

		user, err := m.isValid(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// add user to context for upstream handlers to consume
		ctx := otf.AddUserToContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *authTokenMiddleware) isValid(ctx context.Context, token string) (*otf.User, error) {
	// check if site admin token
	if m.siteToken != "" {
		if m.siteToken == token {
			return &otf.SiteAdmin, nil
		}
	}

	// check if user token
	return m.svc.Get(ctx, otf.UserSpec{AuthenticationToken: &token})
}
