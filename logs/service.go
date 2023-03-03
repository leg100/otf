package logs

import (
	"context"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/rbac"
)

type (
	service interface {
		GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error)
		PutChunk(ctx context.Context, chunk otf.Chunk) error
		tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error)
	}

	Service struct {
		logr.Logger
		otf.PubSubService // subscribe to tail log updates

		run   otf.Authorizer
		proxy db
		db    *pgdb

		api *api
		web *web
	}

	Options struct {
		otf.Authorizer
		otf.Cache
		otf.DB
		*pubsub.Hub
		otf.Verifier
		logr.Logger
		RunAuthorizer otf.Authorizer
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:        opts.Logger,
		PubSubService: opts.Hub,
		proxy: &proxy{
			Logger:        opts.Logger,
			PubSubService: opts.Hub,
			cache:         opts.Cache,
			db:            newPGDB(opts.DB),
		},
		run: opts.RunAuthorizer,
	}
	svc.api = &api{
		service:  &svc,
		Verifier: opts.Verifier,
	}
	svc.web = newWebHandlers(&svc, opts.Logger)

	// Must register table name and service with pubsub broker so that it knows
	// how to lookup chunks in the DB and send them to us via a subscription
	opts.Register("logs", &svc)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// GetByID implements pubsub.Getter
func (s *Service) GetByID(ctx context.Context, chunkID string) (any, error) {
	id, err := strconv.Atoi(chunkID)
	if err != nil {
		return otf.PersistedChunk{}, err
	}
	return s.db.getByID(ctx, id)
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.get(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
		return otf.Chunk{}, nil
	} else if err != nil {
		s.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase.
func (s *Service) PutChunk(ctx context.Context, chunk otf.Chunk) error {
	_, err := s.run.CanAccess(ctx, rbac.PutChunkAction, chunk.RunID)
	if err != nil {
		return err
	}

	persisted, err := s.proxy.put(ctx, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)
		return err
	}
	s.V(2).Info("written logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)

	s.Publish(otf.Event{
		Type:    otf.EventLogChunk,
		Payload: persisted,
	})

	return nil
}

// tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error) {
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
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
	} else if err != nil {
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
	s.V(2).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return ch, nil
}
