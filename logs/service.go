package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/rbac"
)

type (
	LogsService = Service

	Service interface {
		GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error)
		Tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error)
		otf.PutChunkService
	}

	service struct {
		logr.Logger
		otf.PubSubService // subscribe to tail log updates

		run   otf.Authorizer
		proxy chunkdb

		api *api
		web *webHandlers
	}

	chunkdb interface {
		get(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error)
		put(ctx context.Context, opts otf.PutChunkOptions) error
	}

	Options struct {
		logr.Logger
		otf.Cache
		otf.DB
		pubsub.Broker
		otf.Verifier

		RunAuthorizer otf.Authorizer
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:        opts.Logger,
		PubSubService: opts.Broker,
		proxy:         newProxy(opts),
		run:           opts.RunAuthorizer,
	}
	svc.api = &api{
		Verifier: opts.Verifier,
		svc:      &svc,
	}
	svc.web = &webHandlers{
		Logger: opts.Logger,
		svc:    &svc,
	}

	return &svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *service) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.get(ctx, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase
func (s *service) PutChunk(ctx context.Context, opts otf.PutChunkOptions) error {
	_, err := s.run.CanAccess(ctx, rbac.PutChunkAction, opts.RunID)
	if err != nil {
		return err
	}

	if err := s.proxy.put(ctx, opts); err != nil {
		s.Error(err, "writing logs", "id", opts.RunID, "phase", opts.Phase)
		return err
	}
	s.V(2).Info("written logs", "id", opts.RunID, "phase", opts.Phase)

	return nil
}

// tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *service) Tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	subject, err := s.run.CanAccess(ctx, rbac.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}

	// Subscribe first and only then retrieve from DB, guaranteeing that we
	// won't miss any updates
	sub, err := s.Subscribe(ctx, "tail-"+otf.GenerateRandomString(6))
	if err != nil {
		return nil, err
	}

	chunk, err := s.proxy.get(ctx, opts)
	if err != nil {
		s.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
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
				chunk, ok := ev.Payload.(otf.Chunk)
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
					chunk = chunk.Cut(otf.GetChunkOptions{Offset: opts.Offset})
				}
				if len(chunk.Data) == 0 {
					// don't send empty chunks
					continue
				}
				ch <- chunk
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
	s.V(2).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return ch, nil
}
