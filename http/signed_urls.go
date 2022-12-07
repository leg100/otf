package http

import (
	"fmt"
	"net/http"
	"time"
)

// Signer is capable of signing URLs with a limited lifespan.
type Signer interface {
	Sign(string, time.Duration) (string, error)
}

// Verifier is capable of verifying signed URLs
type Verifier interface {
	Verify(string) error
}

// signatureVerifier is middleware that verifies and validates signed URLs
type signatureVerifier struct {
	Verifier
}

func (v *signatureVerifier) handler(next http.Handler) http.Handler {
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

// signedUploadURL creates a signed URL for uploading a configuration version
// blob
func (s *Server) signedUploadURL(cvID string) string {
	url := fmt.Sprintf("/configuration-versions/%s/upload", cvID)
	url, err := s.Sign(url, time.Hour)
	if err != nil {
		panic("signing url: " + url + "; error: " + err.Error())
	}
	return url
}
