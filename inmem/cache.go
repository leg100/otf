package inmem

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/leg100/otf"
)

type CacheConfig struct {
	Size int
	TTL  time.Duration
}

func NewCache(config CacheConfig) (*bigcache.BigCache, error) {
	defaults := bigcache.DefaultConfig(otf.DefaultCacheTTL)

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

	return cache, nil
}
