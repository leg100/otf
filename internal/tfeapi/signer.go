package tfeapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/surl/v2"
)

const (
	SignedPrefixWithSignature = signedPrefix + "/{signature.expiry}"
	signedPrefix              = "/signed"
)

// NewSigner constructs a signer for signing and verifying URLs
func NewSigner(secret []byte) *surl.Signer {
	return surl.New(
		secret,
		surl.PrefixPath(signedPrefix),
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

// verifySignedURL is middleware that verifies signed URLs
func verifySignedURL(v Verifier) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip non-signed URLs
			if !strings.HasPrefix(r.URL.Path, signedPrefix) {
				next.ServeHTTP(w, r)
				return
			}

			if err := v.Verify(r.URL.String()); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
