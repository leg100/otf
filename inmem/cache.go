package inmem

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/leg100/otf"
)

type Cache struct {
	*bigcache.BigCache
}

type CacheConfig struct {
	Size int
	TTL  time.Duration
}

func NewCache(config CacheConfig) (otf.Cache, error) {
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

	return &Cache{BigCache: cache}, nil
}

// GetChunk retrieves a chunk of a cached value.
func (c *Cache) GetChunk(key string, opts otf.GetChunkOptions) ([]byte, error) {
	val, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	if opts.Limit == 0 {
		return val[opts.Offset:], nil
	}

	if opts.Limit > otf.ChunkMaxLimit {
		opts.Limit = otf.ChunkMaxLimit
	}

	// Adjust limit if it extends beyond size of value
	if (opts.Offset + opts.Limit) > len(val) {
		opts.Limit = len(val) - opts.Offset
	}

	return val[opts.Offset:(opts.Offset + opts.Limit)], nil
}

// AppendChunk appends to a cached value. If the key does not exist then a new
// key is created.
func (c *Cache) AppendChunk(key string, chunk []byte) error {
	val, err := c.Get(key)
	if err != nil {
		return c.Set(key, chunk)
	}

	return c.Set(key, append(val, chunk...))
}
