package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/r3labs/sse/v2"
)

type app interface {
	GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error)
	PutChunk(ctx context.Context, chunk Chunk) error
	Tail(ctx context.Context, opts GetChunkOptions) (<-chan otf.Chunk, error)
}

type Application struct {
	otf.Authorizer // authorize access
	logr.Logger
	otf.PubSubService

	proxy ChunkProxy
	db
	*handlers
	*htmlHandlers
}

func NewApplication(opts ApplicationOptions) *Application {
	app := &Application{
		Authorizer:    opts.Authorizer,
		Logger:        opts.Logger,
		PubSubService: opts.PubSubService,
	}

	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	app.handlers = &handlers{
		app:      app,
		Verifier: opts.Verifier,
	}
	app.htmlHandlers = &htmlHandlers{
		app:    app,
		Logger: opts.Logger,
		Server: srv,
	}
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.PubSubService
	otf.Verifier
	logr.Logger
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (a *Application) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	logs, err := a.proxy.GetChunk(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
		return Chunk{}, nil
	} else if err != nil {
		a.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
		return Chunk{}, err
	}
	a.V(2).Info("read logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase.
func (a *Application) PutChunk(ctx context.Context, chunk Chunk) error {
	_, err := a.CanAccessRun(ctx, rbac.PutChunkAction, chunk.RunID)
	if err != nil {
		return err
	}

	persisted, err := a.db.PutChunk(ctx, chunk)
	if err != nil {
		a.Error(err, "writing logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)
		return err
	}
	a.V(2).Info("written logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)

	a.Publish(otf.Event{
		Type:    otf.EventLogChunk,
		Payload: persisted,
	})

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (a *Application) Tail(ctx context.Context, opts GetChunkOptions) (<-chan otf.Chunk, error) {
	subject, err := a.CanAccessRun(ctx, rbac.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}

	// Subscribe first and only then retrieve from DB, guaranteeing that we
	// won't miss any updates
	sub, err := a.Subscribe(ctx, "tail-"+otf.GenerateRandomString(6))
	if err != nil {
		return nil, err
	}

	chunk, err := a.proxy.GetChunk(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
	} else if err != nil {
		a.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
		return nil, err
	}
	opts.Offset += len(chunk.Data)

	ch := make(chan otf.Chunk)
	go func() {
		// send existing chunk
		if len(chunk.Data) > 0 {
			ch <- chunk
		}

		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					close(ch)
					return
				}
				chunk, ok := ev.Payload.(otf.PersistedChunk)
				if !ok {
					// skip non-log events
					continue
				}
				if opts.RunID != chunk.RunID || opts.Phase != chunk.Phase {
					// skip logs for different run/phase
					continue
				}
				if chunk.Offset < opts.Offset {
					// chunk has overlapping offset

					if chunk.Offset+len(chunk.Data) <= opts.Offset {
						// skip entirely overlapping chunk
						continue
					}
					// remove overlapping portion of chunk
					chunk.Chunk = chunk.Cut(otf.GetChunkOptions{Offset: opts.Offset})
				}
				if len(chunk.Data) == 0 {
					// don't send empty chunks
					continue
				}
				ch <- chunk.Chunk
				if chunk.IsEnd() {
					close(ch)
					return
				}
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	a.V(2).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return ch, nil
}
