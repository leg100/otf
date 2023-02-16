package logs

import (
	"context"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/pubsub"
)

type Service struct {
	logr.Logger
	otf.Authorizer // authorize access
	otf.PubSubService

	app *Application
	api *api
	db  *pgdb
	web *web
}

func NewService(opts Options) *Service {
	db := newPGDB(opts.DB)
	app := &Application{
		Logger:        opts.Logger,
		Authorizer:    opts.Authorizer,
		PubSubService: opts.Hub,
		proxy: &proxy{
			Logger:        opts.Logger,
			PubSubService: opts.Hub,
			cache:         opts.Cache,
			db:            newPGDB(opts.DB),
		},
	}
	svc := &Service{
		api: &api{
			app:      app,
			Verifier: opts.Verifier,
		},
		db:  db,
		web: newWebHandlers(app, opts.Logger),
	}

	// Must register table name and service with pubsub broker so that it knows
	// how to lookup chunks in the DB and send them to us via a subscription
	opts.Hub.Register("logs", svc)

	return svc
}

type Options struct {
	otf.Authorizer
	otf.Cache
	otf.DB
	*pubsub.Hub
	otf.Verifier
	logr.Logger
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

// GetByID implements pubsub.Getter
func (s *Service) GetByID(ctx context.Context, chunkID string) (any, error) {
	id, err := strconv.Atoi(chunkID)
	if err != nil {
		return PersistedChunk{}, err
	}
	return s.db.getByID(ctx, id)
}

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) GetChunk(ctx context.Context, opts GetChunkOptions) (Chunk, error) {
	return s.app.get(ctx, opts)
}

// PutChunk writes a chunk of logs for a phase.
func (s *Service) PutChunk(ctx context.Context, chunk Chunk) error {
	return s.app.put(ctx, chunk)
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (s *Service) Tail(ctx context.Context, opts GetChunkOptions) (<-chan Chunk, error) {
	return s.app.tail(ctx, opts)
}
