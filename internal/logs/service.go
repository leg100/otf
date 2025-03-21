package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type (
	Service struct {
		logr.Logger
		authz.Interface

		api    *api
		web    *webHandlers
		broker pubsub.SubscriptionService[Chunk]

		chunkproxy
	}

	chunkproxy interface {
		Start(ctx context.Context) error
		get(ctx context.Context, opts GetChunkOptions) (Chunk, error)
		put(ctx context.Context, chunk Chunk) error
	}

	Options struct {
		logr.Logger
		internal.Cache
		*sql.DB
		*sql.Listener
		internal.Verifier

		Authorizer *authz.Authorizer
	}
)

func NewService(opts Options) *Service {
	db := &pgdb{opts.DB}
	svc := Service{
		Logger:    opts.Logger,
		Interface: opts.Authorizer,
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
		func(ctx context.Context, chunkID resource.TfeID, action sql.Action) (Chunk, error) {
			if action == sql.DeleteAction {
				return Chunk{TfeID: chunkID}, nil
			}
			return db.getChunk(ctx, chunkID)
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

func (s *Service) WatchLogs(ctx context.Context) (<-chan pubsub.Event[Chunk], func()) {
	return s.broker.Subscribe(ctx)
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	logs, err := s.chunkproxy.get(ctx, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset)
		return Chunk{}, err
	}
	s.V(9).Info("read logs", "id", opts.RunID, "offset", opts.Offset)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase
func (s *Service) PutChunk(ctx context.Context, opts PutChunkOptions) error {
	_, err := s.Authorize(ctx, authz.PutChunkAction, &authz.AccessRequest{ID: &opts.RunID})
	if err != nil {
		return err
	}

	chunk, err := newChunk(opts)
	if err != nil {
		s.Error(err, "creating log chunk", "run_id", opts, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	if err := s.put(ctx, chunk); err != nil {
		s.Error(err, "writing logs", "chunk_id", chunk.TfeID, "run_id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	s.V(3).Info("written logs", "id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) Tail(ctx context.Context, opts GetChunkOptions) (<-chan Chunk, error) {
	subject, err := s.Authorize(ctx, authz.TailLogsAction, &authz.AccessRequest{ID: &opts.RunID})
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
	relay := make(chan Chunk)
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
				chunk = chunk.Cut(GetChunkOptions{Offset: opts.Offset})
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
