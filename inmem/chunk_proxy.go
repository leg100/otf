package inmem

import (
	"context"

	"github.com/leg100/otf"
)

// ChunkProxy is a caching proxy for log chunks, proxying requests to the
// backend.
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
func (c *ChunkProxy) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	// Try the cache first
	if data, err := c.cache.Get(otf.LogCacheKey(opts.RunID, opts.Phase)); err == nil {
		return otf.Chunk{Data: data}.Cut(opts), nil
	}

	// Fall back to getting chunk from backend
	chunk, err := c.backend.GetChunk(ctx, opts)
	if err != nil {
		return otf.Chunk{}, err
	}
	// Cache it
	if err := c.cache.Set(chunk.Key(), chunk.Data); err != nil {
		return otf.Chunk{}, err
	}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

func (c *ChunkProxy) GetChunkByID(ctx context.Context, chunkID int) (otf.PersistedChunk, error) {
	return c.backend.GetChunkByID(ctx, chunkID)
}

// PutChunk writes a chunk of data to the backend store before caching it.
func (c *ChunkProxy) PutChunk(ctx context.Context, chunk otf.Chunk) (otf.PersistedChunk, error) {
	// Write to backend
	persisted, err := c.backend.PutChunk(ctx, chunk)
	if err != nil {
		return otf.PersistedChunk{}, err
	}

	// First chunk can safely be written straight to cache
	if chunk.IsStart() {
		return persisted, c.cache.Set(chunk.Key(), chunk.Data)
	}

	// Otherwise, append chunk to cache
	if previous, err := c.cache.Get(chunk.Key()); err == nil {
		return persisted, c.cache.Set(chunk.Key(), append(previous, chunk.Data...))
	}

	// Cache needs re-populating from store
	all, err := c.backend.GetChunk(ctx, otf.GetChunkOptions{
		RunID: chunk.RunID,
		Phase: chunk.Phase,
	})
	if err == nil {
		return persisted, c.cache.Set(chunk.Key(), all.Data)
	} else {
		return persisted, err
	}
}
