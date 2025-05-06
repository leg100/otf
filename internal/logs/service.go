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
		*authz.Authorizer

		api    *api
		web    *webHandlers
		broker pubsub.SubscriptionService[Chunk]
		db     *pgdb
		tailer *tailer
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
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         db,
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
				return Chunk{ID: chunkID}, nil
			}
			return db.getChunk(ctx, chunkID)
		},
	)
	svc.tailer = &tailer{
		broker: svc.broker,
		client: &svc,
	}
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *Service) GetLogs(ctx context.Context, runID resource.TfeID, phase internal.PhaseType) ([]byte, error) {
	logs, err := s.db.getLogs(ctx, runID, phase)
	if err != nil {
		s.Error(err, "reading all logs", "run_id", runID, "phase", phase)
		return nil, err
	}
	s.V(9).Info("read all logs", "run_id", runID, "phase", phase)
	return logs, nil
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	logs, err := s.db.getLogs(ctx, opts.RunID, opts.Phase)
	if err != nil {
		s.Error(err, "retrieving log chunk", "id", opts.RunID, "offset", opts.Offset)
		return Chunk{}, err
	}
	s.V(9).Info("retrieved log chunk", "id", opts.RunID, "offset", opts.Offset)
	chunk := Chunk{RunID: opts.RunID, Phase: opts.Phase, Data: logs}
	// Cut chunk down to requested size.
	return chunk.Cut(opts), nil
}

// PutChunk writes a chunk of logs for a phase
func (s *Service) PutChunk(ctx context.Context, opts PutChunkOptions) error {
	_, err := s.Authorize(ctx, authz.PutChunkAction, opts.RunID)
	if err != nil {
		return err
	}

	chunk, err := newChunk(opts)
	if err != nil {
		s.Error(err, "creating log chunk", "run_id", opts, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	if err := s.db.put(ctx, chunk); err != nil {
		s.Error(err, "writing logs", "chunk_id", chunk.ID, "run_id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)
		return err
	}
	s.V(3).Info("written logs", "id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset)

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) Tail(ctx context.Context, opts GetChunkOptions) (<-chan Chunk, error) {
	subject, err := s.Authorize(ctx, authz.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}
	tail, err := s.tailer.Tail(ctx, opts)
	if err != nil {
		s.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
		return nil, err
	}
	s.V(9).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return tail, nil
}

func (s *Service) WatchLogs(ctx context.Context) (<-chan pubsub.Event[Chunk], func()) {
	return s.broker.Subscribe(ctx)
}
