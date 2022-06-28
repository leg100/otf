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
func (c *ChunkProxy) GetChunk(ctx context.Context, id string, phase otf.PhaseType, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// Try the cache first
	if chunk, err := c.cache.Get(otf.LogCacheKey(id, phase)); err == nil {
		return otf.UnmarshalChunk(chunk).Cut(opts)
	}

	// Fall back to getting chunk from backend
	chunk, err := c.backend.GetChunk(ctx, id, phase, otf.GetChunkOptions{})
	if err != nil {
		return otf.Chunk{}, err
	}

	// Cache it
	if err := c.cache.Set(otf.LogCacheKey(id, phase), chunk.Marshal()); err != nil {
		return otf.Chunk{}, err
	}

	// Cut chunk down to requested size.
	return chunk.Cut(opts)
}

// PutChunk writes a chunk of data to the backend store before caching it.
func (c *ChunkProxy) PutChunk(ctx context.Context, id string, phase otf.PhaseType, chunk otf.Chunk) error {
	// Write to backend
	if err := c.backend.PutChunk(ctx, id, phase, chunk); err != nil {
		return err
	}

	// First chunk can safely be written straight to cache
	if chunk.Start {
		return c.cache.Set(otf.LogCacheKey(id, phase), chunk.Marshal())
	}

	// Otherwise, append chunk to cache
	if previous, err := c.cache.Get(otf.LogCacheKey(id, phase)); err == nil {
		return c.cache.Set(otf.LogCacheKey(id, phase), append(previous, chunk.Marshal()...))
	}

	// Cache needs re-populating from store
	if all, err := c.backend.GetChunk(ctx, id, phase, otf.GetChunkOptions{}); err == nil {
		return c.cache.Set(otf.LogCacheKey(id, phase), all.Marshal())
	} else {
		return err
	}
}
