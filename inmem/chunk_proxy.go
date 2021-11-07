package inmem

import (
	"context"

	"github.com/leg100/otf"
)

type ChunkProxy struct {
	cache   otf.Cache
	backend otf.ChunkStore
}

func NewChunkProxy(cache otf.Cache, backend otf.ChunkStore) (otf.ChunkStore, error) {
	return &ChunkProxy{
		cache:   cache,
		backend: backend,
	}, nil
}

// GetChunk attempts to retrieve a chunk from the cache before falling back to
// using the backend store.
func (c *ChunkProxy) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error) {
	if all, err := c.cache.Get(otf.LogCacheKey(id)); err == nil {
		return otf.GetChunk(all, opts)
	}

	// Get all chunks from backend
	all, err := c.backend.GetChunk(ctx, id, otf.GetChunkOptions{})
	if err != nil {
		return nil, err
	}

	// Populate cache
	if err := c.cache.Set(otf.LogCacheKey(id), all); err != nil {
		return nil, err
	}

	// Return requested chunk
	return otf.GetChunk(all, opts)
}

// PutChunk writes a chunk of data to the backend store before caching it.
func (c *ChunkProxy) PutChunk(ctx context.Context, key string, chunk []byte, opts otf.PutChunkOptions) error {
	if err := c.backend.PutChunk(ctx, key, chunk, opts); err != nil {
		return err
	}

	// First chunk can safely be written straight to cache
	if opts.Start {
		return c.cache.Set(otf.LogCacheKey(key), chunk)
	}

	// Otherwise, append chunk to cache
	if previous, err := c.cache.Get(otf.LogCacheKey(key)); err == nil {
		return c.cache.Set(otf.LogCacheKey(key), append(previous, chunk...))
	}

	// Cache needs re-populating from store
	all, err := c.backend.GetChunk(ctx, key, otf.GetChunkOptions{})
	if err != nil {
		return err
	}
	if err := c.cache.Set(otf.LogCacheKey(key), all); err != nil {
		return err
	}

	return nil
}
