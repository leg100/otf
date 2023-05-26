package inmem

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/leg100/otf/internal"
)

type CacheConfig struct {
	// Total capacity of cache in MB.
	Size int
	// Time-to-live for each cache entry before automatic deletion.
	TTL time.Duration
}

func NewCache(config CacheConfig) (*bigcache.BigCache, error) {
	defaults := bigcache.DefaultConfig(internal.DefaultCacheTTL)
	defaults.Verbose = false

	if config.TTL != 0 {
		defaults.LifeWindow = config.TTL
	}

	if config.Size != 0 {
		defaults.HardMaxCacheSize = config.Size / defaults.Shards
	}

	cache, err := bigcache.NewBigCache(defaults)
	if err != nil {
		return nil, err
	}

	// continuously gather metrics
	cacheSize.Set(float64(config.Size))
	go func() {
		for range time.Tick(5 * time.Second) {
			cacheUsed.Set(float64(cache.Capacity()))
			cacheEntries.Set(float64(cache.Len()))

			stats := cache.Stats()
			cacheHits.Set(float64(stats.Hits))
			cacheMisses.Set(float64(stats.Misses))
			cacheDelHits.Set(float64(stats.DelHits))
			cacheDelMisses.Set(float64(stats.DelMisses))
		}
	}()

	return cache, nil
}
