/*
Package app implements services, co-ordinating between the layers of the project.
*/
package app

import (
	"context"
	"fmt"

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

var _ otf.Application = (*Application)(nil)

// Application encompasses services for interacting between components of the
// otf server
type Application struct {
	db     otf.DB
	cache  otf.Cache
	proxy  otf.ChunkStore
	queues *inmem.WorkspaceQueueManager
	latest *inmem.LatestRunManager

	*otf.RunFactory
	*otf.WorkspaceFactory
	Mapper
	otf.PubSubService
	logr.Logger
}

// NewApplication constructs an application, initialising various services and
// daemons.
func NewApplication(ctx context.Context, logger logr.Logger, db otf.DB, cache *bigcache.BigCache, pubsub otf.PubSubService) (*Application, error) {
	// Setup ID mapper
	mapper := inmem.NewMapper()

	app := &Application{
		PubSubService: pubsub,
		Mapper:        mapper,
		cache:         cache,
		db:            db,
		Logger:        logger,
	}
	app.WorkspaceFactory = &otf.WorkspaceFactory{OrganizationService: app}
	app.RunFactory = &otf.RunFactory{
		WorkspaceService:            app,
		ConfigurationVersionService: app,
	}

	// Setup latest run manager
	latest, err := inmem.NewLatestRunManager(ctx, app)
	if err != nil {
		return nil, err
	}
	app.latest = latest

	proxy, err := inmem.NewChunkProxy(cache, db)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	app.proxy = proxy

	// Populate mappings with identifiers
	if err := mapper.Populate(ctx, app, app); err != nil {
		return nil, err
	}

	queues := inmem.NewWorkspaceQueueManager()
	if err := queues.Populate(ctx, app); err != nil {
		return nil, fmt.Errorf("populating workspace queues: %w", err)
	}
	app.queues = queues

	return app, nil
}

// Tx provides a callback in which all db interactions are wrapped within a
// transaction. Useful for ensuring multiple service calls succeed together.
func (a *Application) Tx(ctx context.Context, tx func(a *Application) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		// make a copy of the app and assign a db tx wrapper
		appTx := &Application{
			PubSubService:    a.PubSubService,
			Mapper:           a.Mapper,
			cache:            a.cache,
			Logger:           a.Logger,
			WorkspaceFactory: a.WorkspaceFactory,
			RunFactory:       a.RunFactory,
			latest:           a.latest,
			proxy:            a.proxy,
			queues:           a.queues,
			db:               db,
		}
		return tx(appTx)
	})
}

// WithLock provides a callback in which the application's database connection
// possesses a lock with the given ID, with the guarantee that that lock is
// exclusive, i.e. no other connection has a lock with the same ID. If a lock
// with the given ID is already present then this method will block until it is
// released.
func (a *Application) WithLock(ctx context.Context, id int64, cb func(otf.Application) error) error {
	return a.db.WaitAndLock(ctx, id, func(db otf.DB) error {
		// make a copy of the app and assign a db wrapped with a session-lock
		appWithLock := &Application{
			PubSubService:    a.PubSubService,
			Mapper:           a.Mapper,
			cache:            a.cache,
			Logger:           a.Logger,
			WorkspaceFactory: a.WorkspaceFactory,
			RunFactory:       a.RunFactory,
			latest:           a.latest,
			proxy:            a.proxy,
			queues:           a.queues,
			db:               db,
		}
		return cb(appWithLock)
	})
}
