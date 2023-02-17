package logs

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// cacheKey generates a key for caching log chunks.
func cacheKey(runID string, phase otf.PhaseType) string {
	return fmt.Sprintf("%s.%s.log", runID, string(phase))
}

// proxy is a caching proxy for log chunks
type proxy struct {
	cache otf.Cache
	db    db

	otf.PubSubService
	logr.Logger
}

// Start chunk proxy daemon, which keeps the cache up-to-date with logs
// published on other nodes in the cluster
func (c *proxy) Start(ctx context.Context) error {
	ch := make(chan otf.Chunk)
	defer close(ch)

	// TODO: if it loses its connection to the stream it should keep retrying,
	// with a backoff alg, and it should invalidate the cache *entirely* because
	// it may have missed updates, potentially rendering the cache stale.
	sub, err := c.Subscribe(ctx, "chunk-proxy")
	if err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-sub:
			if !ok {
				return nil
			}
			chunk, ok := event.Payload.(PersistedChunk)
			if !ok {
				// skip non-log events
				continue
			}
			if err := c.cacheChunk(ctx, chunk.Chunk); err != nil {
				c.Error(err, "caching log chunk")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetChunk attempts to retrieve a chunk from the cache before falling back to
// using the backend store.
func (c *proxy) get(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	key := cacheKey(opts.RunID, opts.Phase)

	// Try the cache first
	if data, err := c.cache.Get(key); err == nil {
		return Chunk{Data: data}.Cut(opts), nil
	}
	// Fall back to getting chunk from backend
	chunk, err := c.db.get(ctx, opts)
	if err != nil {
		return Chunk{}, err
	}
	// Cache it
	if err := c.cache.Set(key, chunk.Data); err != nil {
		return Chunk{}, err
	}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

// PutChunk writes a chunk of data to the backend store before caching it.
func (c *proxy) put(ctx context.Context, chunk Chunk) (PersistedChunk, error) {
	// Write to backend
	persisted, err := c.db.put(ctx, chunk)
	if err != nil {
		return PersistedChunk{}, err
	}

	// Then cache it
	if err := c.cacheChunk(ctx, persisted.Chunk); err != nil {
		return PersistedChunk{}, err
	}

	return persisted, nil
}

func (c *proxy) cacheChunk(ctx context.Context, chunk Chunk) error {
	key := cacheKey(chunk.RunID, chunk.Phase)

	// first chunk: don't append
	if chunk.IsStart() {
		return c.cache.Set(key, chunk.Data)
	}
	// successive chunks: append
	if previous, err := c.cache.Get(key); err == nil {
		return c.cache.Set(key, append(previous, chunk.Data...))
	}
	// no cache entry; repopulate cache from db
	all, err := c.db.get(ctx, GetChunkOptions{
		RunID: chunk.RunID,
		Phase: chunk.Phase,
	})
	if err != nil {
		return err
	}
	return c.cache.Set(key, all.Data)
}
