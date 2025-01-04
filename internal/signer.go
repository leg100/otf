package internal

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/surl/v2"
)

// NewSigner constructs a signer for signing and verifying URLs
func NewSigner(secret []byte) *surl.Signer {
	return surl.New(secret,
		surl.PrefixPath("/signed"),
		surl.WithPathFormatter(),
		surl.WithBase58Expiry(),
		surl.SkipQuery(),
	)
}

// Signer cryptographically signs URLs with a limited lifespan.
type Signer interface {
	Sign(string, time.Time) (string, error)
}

// Verifier verifies signed URLs
type Verifier interface {
	Verify(string) error
}

// VerifySignedURL is middleware that verifies signed URLs
func VerifySignedURL(v Verifier) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := v.Verify(r.URL.String()); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
