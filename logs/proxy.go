package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// ChunkProxy is a caching proxy for log chunks, proxying requests to the
// backend.
type ChunkProxy struct {
	cache   otf.Cache
	backend ChunkStore

	otf.Application
	logr.Logger
}

func NewChunkProxy(app otf.Application, logger logr.Logger, cache otf.Cache, backend ChunkStore) (*ChunkProxy, error) {
	return &ChunkProxy{
		Application: app,
		Logger:      logger.WithValues("component", "chunk_proxy"),
		cache:       cache,
		backend:     backend,
	}, nil
}

// Start the chunk proxy daemon, which keeps the cache up-to-date.
//
func (c *ChunkProxy) Start(ctx context.Context) error {
	// TODO: if it loses its connection to the stream it should keep retrying,
	// with a backoff alg, and it should invalidate the cache *entirely* because
	// it may have missed up dates, potentially rendering the cache stale.
	sub, err := c.WatchLogs(ctx, otf.WatchLogsOptions{Name: otf.String("chunk-proxy")})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-sub:
			if !ok {
				return nil
			}
			if err := c.cacheChunk(ctx, chunk); err != nil {
				c.Error(err, "caching log chunk")
			}
		}
	}
}

// GetChunk attempts to retrieve a chunk from the cache before falling back to
// using the backend store.
func (c *ChunkProxy) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	// Try the cache first
	if data, err := c.cache.Get(otf.LogCacheKey(opts.RunID, opts.Phase)); err == nil {
		return Chunk{Data: data}.Cut(opts), nil
	}

	// Fall back to getting chunk from backend
	chunk, err := c.backend.GetChunk(ctx, opts)
	if err != nil {
		return Chunk{}, err
	}
	// Cache it
	if err := c.cache.Set(chunk.Key(), chunk.Data); err != nil {
		return Chunk{}, err
	}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

func (c *ChunkProxy) GetChunkByID(ctx context.Context, chunkID int) (PersistedChunk, error) {
	return c.backend.GetChunkByID(ctx, chunkID)
}

// PutChunk writes a chunk of data to the backend store before caching it.
func (c *ChunkProxy) PutChunk(ctx context.Context, chunk Chunk) (PersistedChunk, error) {
	// Write to backend
	persisted, err := c.backend.PutChunk(ctx, chunk)
	if err != nil {
		return PersistedChunk{}, err
	}

	// Then cache it
	if err := c.cacheChunk(ctx, persisted.Chunk); err != nil {
		return PersistedChunk{}, err
	}

	return persisted, nil
}

func (c *ChunkProxy) cacheChunk(ctx context.Context, chunk Chunk) error {
	// First chunk can safely be written straight to cache
	if chunk.IsStart() {
		return c.cache.Set(chunk.Key(), chunk.Data)
	}

	// Otherwise, append chunk to cache
	if previous, err := c.cache.Get(chunk.Key()); err == nil {
		return c.cache.Set(chunk.Key(), append(previous, chunk.Data...))
	}

	// Uncached; cache needs re-populating from store
	all, err := c.backend.GetChunk(ctx, GetChunkOptions{
		RunID: chunk.RunID,
		Phase: chunk.Phase,
	})
	if err != nil {
		return err
	}
	return c.cache.Set(chunk.Key(), all.Data)
}
