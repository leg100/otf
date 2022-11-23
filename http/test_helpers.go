package http

import (
	"net/http"

	"github.com/leg100/otf"
)

// allowAllMiddleware by-passes authz for testing puroses.
func allowAllMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otf.AddSubjectToContext(r.Context(), &otf.Superuser{})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
