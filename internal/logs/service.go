package logs

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		api    *api
		web    *webHandlers
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
	svc.tailer = &tailer{
		broker: pubsub.NewBroker[Chunk](
			opts.Logger,
			opts.Listener,
			"chunks",
		),
		client: &svc,
	}
	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// GetChunk retrieves a chunk of logs for a run phase.
func (s *Service) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	chunk, err := s.db.getChunk(ctx, opts)
	if err != nil {
		s.Error(err, "retrieving log chunk", "run_id", opts.RunID, "phase", opts.Phase, "offset", opts.Offset, "limit", opts.Limit)
		return Chunk{}, err
	}
	s.V(9).Info("retrieved log chunk", "chunk", chunk)
	return chunk, nil
}

// PutChunk writes a chunk of logs for a run phase
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
	if err := s.db.putChunk(ctx, chunk); err != nil {
		s.Error(err, "writing log chunk", "chunk", chunk)
		return err
	}
	s.V(3).Info("written log chunk", "chunk", chunk)

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) Tail(ctx context.Context, opts TailOptions) (<-chan Chunk, error) {
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
