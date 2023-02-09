package otf

import "time"

// Signer cryptographically signs strings with a limited lifespan.
type Signer interface {
	Sign(string, time.Duration) (string, error)
}

// Verifier is capable of verifying signed URLs
type Verifier interface {
	Verify(string) error
}
