package http

import (
	"net/http"

	"github.com/leg100/otf"
)

// SignatureVerifier is middleware that verifies signed URLs
type SignatureVerifier struct {
	otf.Verifier
}

func (v *SignatureVerifier) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := v.Verify(r.URL.String()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
