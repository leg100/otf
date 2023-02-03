package otf

import "time"

// Signer cryptographically signs strings with a limited lifespan.
type Signer interface {
	Sign(string, time.Duration) (string, error)
}
