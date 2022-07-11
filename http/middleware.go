package http

import (
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
			http.Error(w, "malformed token", http.StatusUnauthorized)
			return
		}
		token := hdr[1]

		// check if it is a site token
		if m.siteToken != "" {
			if m.siteToken == token {
				next.ServeHTTP(w, r)
				return
			}
		}

		// check if user token
		user, err := m.svc.Get(r.Context(), otf.UserSpec{AuthenticationToken: &token})
		if err == otf.ErrResourceNotFound {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add user to context for upstream handlers to consume
		ctx := addUserToContext(r.Context(), user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
