package otf

import (
	"time"

	"github.com/leg100/surl"
)

// NewSigner constructs a signer for signing and verifying URLs
func NewSigner(secret string) *surl.Signer {
	return surl.New([]byte(secret),
		surl.PrefixPath("/signed"),
		surl.WithPathFormatter(),
		surl.WithBase58Expiry(),
		surl.SkipQuery(),
	)
}

// Signer cryptographically signs URLs with a limited lifespan.
type Signer interface {
	Sign(string, time.Duration) (string, error)
}

// Verifier verifies signed URLs
type Verifier interface {
	Verify(string) error
}
