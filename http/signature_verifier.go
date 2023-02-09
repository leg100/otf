package http

import (
	"fmt"
	"net/http"
	"time"

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

// signedLogURL creates a signed URL for retrieving logs for a run phase.
func (s *Server) signedLogURL(r *http.Request, runID, phase string) string {
	url := fmt.Sprintf("/runs/%s/logs/%s", runID, phase)
	url, err := s.Sign(url, time.Hour)
	if err != nil {
		panic("signing url: " + url + "; error: " + err.Error())
	}
	// Terraform CLI expects an absolute URL
	return Absolute(r, url)
}
