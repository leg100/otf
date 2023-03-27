package otf

import (
	"time"
)

// DefaultCacheTTL is the default TTL for cached objects
var DefaultCacheTTL = 10 * time.Minute

// Cache is a key-value cache.
type Cache interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}
