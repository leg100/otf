package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
)

type (
	//Service interface {
	//	GetChunk(ctx context.Context, opts internal.GetChunkOptions) (internal.Chunk, error)
	//	Tail(ctx context.Context, opts internal.GetChunkOptions) (<-chan internal.Chunk, error)
	//	WatchLogs(ctx context.Context) (<-chan pubsub.Event[internal.Chunk], func())
	//	internal.PutChunkService
	//	Start(context.Context) error
	//}

	Service struct {
		logr.Logger

		run internal.Authorizer

		api    *api
		web    *webHandlers
		broker pubsub.SubscriptionService[internal.Chunk]

		chunkproxy
	}

	chunkproxy interface {
		Start(ctx context.Context) error
		get(ctx context.Context, opts internal.GetChunkOptions) (internal.Chunk, error)
		put(ctx context.Context, opts internal.PutChunkOptions) error
	}

	Options struct {
		logr.Logger
		internal.Cache
		*sql.DB
		*sql.Listener
		internal.Verifier

		RunAuthorizer internal.Authorizer
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger: opts.Logger,
		run:    opts.RunAuthorizer,
	}
	svc.api = &api{
		Verifier: opts.Verifier,
		svc:      &svc,
	}
	svc.web = &webHandlers{
		Logger: opts.Logger,
		svc:    &svc,
	}
	svc.broker = pubsub.NewBroker(
		opts.Logger,
		opts.Listener,
		"logs",
		func(ctx context.Context, id string, action sql.Action) (internal.Chunk, error) {
			if action == sql.DeleteAction {
				return internal.Chunk{ID: id}, nil
			}
			return db.getChunk(ctx, id)
		},
	)
	svc.chunkproxy = &proxy{
		Logger: opts.Logger,
		cache:  opts.Cache,
		db:     db,
		broker: svc.broker,
	}
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *Service) WatchLogs(ctx context.Context) (<-chan pubsub.Event[internal.Chunk], func()) {
	return s.broker.Subscribe(ctx)
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) GetChunk(ctx context.Context, opts internal.GetChunkOptions) (internal.Chunk, error) {
	logs, err := s.chunkproxy.get(ctx, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset)
		return internal.Chunk{}, err
	}
	s.V(9).Info("read logs", "id", opts.RunID, "offset", opts.Offset)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase
func (s *Service) PutChunk(ctx context.Context, opts internal.PutChunkOptions) error {
	_, err := s.run.CanAccess(ctx, rbac.PutChunkAction, opts.RunID)
	if err != nil {
		return err
	}

	if err := s.chunkproxy.put(ctx, opts); err != nil {
		s.Error(err, "writing logs", "id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	s.V(3).Info("written logs", "id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) Tail(ctx context.Context, opts internal.GetChunkOptions) (<-chan internal.Chunk, error) {
	subject, err := s.run.CanAccess(ctx, rbac.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}

	// Subscribe first and only then retrieve from DB, guaranteeing that we
	// won't miss any updates
	sub, _ := s.broker.Subscribe(ctx)

	chunk, err := s.chunkproxy.get(ctx, opts)
	if err != nil {
		s.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
		return nil, err
	}
	opts.Offset += len(chunk.Data)

	// relay is the chan returned to the caller on which chunks are relayed to.
	relay := make(chan internal.Chunk)
	go func() {
		// send existing chunk
		if len(chunk.Data) > 0 {
			relay <- chunk
		}

		// relay chunks from subscription
		for ev := range sub {
			chunk := ev.Payload
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
				chunk = chunk.Cut(internal.GetChunkOptions{Offset: opts.Offset})
			}
			if len(chunk.Data) == 0 {
				// don't send empty chunks
				continue
			}
			relay <- chunk
			if chunk.IsEnd() {
				break
			}
		}
		close(relay)
	}()
	s.V(9).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return relay, nil
}
