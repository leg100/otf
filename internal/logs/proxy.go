package logs

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
)

type (
	// proxy is a caching proxy for log chunks
	proxy struct {
		cache  internal.Cache
		db     proxydb
		broker pubsub.SubscriptionService[Chunk]

		logr.Logger
	}

	proxydb interface {
		getAllLogs(ctx context.Context, runID resource.ID, phase internal.PhaseType) ([]byte, error)
		put(ctx context.Context, chunk Chunk) error
	}
)

// Start chunk proxy daemon, which keeps the cache up-to-date with logs
// published across the cluster.
func (p *proxy) Start(ctx context.Context) error {
	// TODO: if it loses its connection to the stream it should keep retrying,
	// with a backoff alg, and it should invalidate the cache *entirely* because
	// it may have missed updates, potentially rendering the cache stale.
	sub, unsub := p.broker.Subscribe(ctx)
	defer unsub()

	for event := range sub {
		chunk := event.Payload
		key := cacheKey(chunk.RunID, chunk.Phase)

		var logs []byte
		// The first log chunk can be written straight to the cache, whereas
		// successive chunks require the cache to be checked first.
		if chunk.IsStart() {
			logs = chunk.Data
		} else {
			if existing, err := p.cache.Get(key); err != nil {
				// no cache entry; retrieve logs from db
				logs, err = p.db.getAllLogs(ctx, chunk.RunID, chunk.Phase)
				if err != nil {
					return err
				}
			} else {
				// append received chunk to existing cached logs
				logs = append(existing, chunk.Data...)
			}
		}
		if err := p.cache.Set(key, logs); err != nil {
			p.Error(err, "caching log chunk")
		}
	}
	return pubsub.ErrSubscriptionTerminated
}

// GetChunk attempts to retrieve a chunk from the cache before falling back to
// using the backend store.
func (p *proxy) get(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	key := cacheKey(opts.RunID, opts.Phase)

	data, err := p.cache.Get(key)
	if err != nil {
		// fall back to retrieving from db...
		data, err = p.db.getAllLogs(ctx, opts.RunID, opts.Phase)
		if err != nil {
			return Chunk{}, err
		}
		// ...and cache it
		if err := p.cache.Set(key, data); err != nil {
			p.Error(err, "caching log chunk")
		}
	}
	chunk := Chunk{RunID: opts.RunID, Phase: opts.Phase, Data: data}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

// put writes a chunk of data to the db
func (p *proxy) put(ctx context.Context, chunk Chunk) error {
	// db triggers an event, which proxy listens for to populate its cache
	return p.db.put(ctx, chunk)
}

// cacheKey generates a key for caching log chunks.
func cacheKey(runID resource.ID, phase internal.PhaseType) string {
	return fmt.Sprintf("%s.%s.log", runID, string(phase))
}
