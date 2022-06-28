package otf

import (
	"fmt"
	"time"
)

// DefaultCacheTTL is the default TTL for cached objects
var DefaultCacheTTL = 10 * time.Minute

// Cache is key-value cache, for performance purposes.
type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte) error
}

// Funcs for generating unique keys for cache entries.

func ConfigVersionCacheKey(id string) string { return fmt.Sprintf("%s.tar.gz", id) }
func StateVersionCacheKey(id string) string  { return fmt.Sprintf("%s.json", id) }
func LogCacheKey(runID string, phase PhaseType) string {
	return fmt.Sprintf("%s.%s.log", runID, string(phase))
}
