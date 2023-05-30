package logs

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
)

type (
	// proxy is a caching proxy for log chunks
	proxy struct {
		cache internal.Cache
		db    db

		pubsub.PubSubService
		logr.Logger
	}

	db interface {
		GetLogs(ctx context.Context, runID string, phase internal.PhaseType) ([]byte, error)
		put(ctx context.Context, opts internal.PutChunkOptions) (string, error)
	}
)

func newProxy(opts Options) *proxy {
	db := &pgdb{opts.DB}
	p := &proxy{
		Logger:        opts.Logger,
		PubSubService: opts.Broker,
		cache:         opts.Cache,
		db:            db,
	}
	// Register with broker so that it can relay log chunks
	opts.Register("logs", db)
	return p
}

// Start chunk proxy daemon, which keeps the cache up-to-date with logs
// published across the cluster.
func (p *proxy) Start(ctx context.Context) error {
	// TODO: if it loses its connection to the stream it should keep retrying,
	// with a backoff alg, and it should invalidate the cache *entirely* because
	// it may have missed updates, potentially rendering the cache stale.
	sub, err := p.Subscribe(ctx, "chunk-proxy-")
	if err != nil {
		return err
	}

	for event := range sub {
		chunk, ok := event.Payload.(internal.Chunk)
		if !ok {
			continue
		}
		key := cacheKey(chunk.RunID, chunk.Phase)

		var logs []byte
		// The first log chunk can be written straight to the cache, whereas
		// successive chunks require the cache to be checked first.
		if chunk.IsStart() {
			logs = chunk.Data
		} else {
			if existing, err := p.cache.Get(key); err != nil {
				// no cache entry; retrieve logs from db
				logs, err = p.db.GetLogs(ctx, chunk.RunID, chunk.Phase)
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
	return nil
}

// GetChunk attempts to retrieve a chunk from the cache before falling back to
// using the backend store.
func (p *proxy) get(ctx context.Context, opts internal.GetChunkOptions) (internal.Chunk, error) {
	key := cacheKey(opts.RunID, opts.Phase)

	data, err := p.cache.Get(key)
	if err != nil {
		// fall back to retrieving from db...
		data, err = p.db.GetLogs(ctx, opts.RunID, opts.Phase)
		if err != nil {
			return internal.Chunk{}, err
		}
		// ...and cache it
		if err := p.cache.Set(key, data); err != nil {
			p.Error(err, "caching log chunk")
		}
	}
	chunk := internal.Chunk{RunID: opts.RunID, Phase: opts.Phase, Data: data}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

// put writes a chunk of data to the db
func (p *proxy) put(ctx context.Context, opts internal.PutChunkOptions) error {
	id, err := p.db.put(ctx, opts)
	if err != nil {
		return err
	}
	// make a chunk from the options and the id
	chunk := internal.Chunk{
		ID:     id,
		RunID:  opts.RunID,
		Phase:  opts.Phase,
		Data:   opts.Data,
		Offset: opts.Offset,
	}
	// publish chunk for caching
	p.Publish(pubsub.Event{Type: pubsub.EventLogChunk, Payload: chunk})
	return nil
}

// cacheKey generates a key for caching log chunks.
func cacheKey(runID string, phase internal.PhaseType) string {
	return fmt.Sprintf("%s.%s.log", runID, string(phase))
}
