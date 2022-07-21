/*
Package app implements services, co-ordinating between the layers of the project.
*/
package app

import (
	"fmt"

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

var (
	_ otf.Application = (*Application)(nil)
)

type Application struct {
	db     otf.DB
	cache  otf.Cache
	proxy  otf.ChunkStore
	queues *inmem.WorkspaceQueueManager
	latest *inmem.LatestRunManager

	*otf.RunFactory
	*otf.WorkspaceFactory
	*inmem.Mapper
	otf.EventService
	logr.Logger
}

func NewApplication(logger logr.Logger, db otf.DB, cache *bigcache.BigCache) (*Application, error) {
	// Setup event broker
	events := inmem.NewEventService(logger)

	// Setup ID mapper
	mapper := inmem.NewMapper()

	app := &Application{
		EventService: events,
		Mapper:       mapper,
		cache:        cache,
		db:           db,
		Logger:       logger,
	}
	app.WorkspaceFactory = &otf.WorkspaceFactory{OrganizationService: app}
	app.RunFactory = &otf.RunFactory{
		WorkspaceService:            app,
		ConfigurationVersionService: app,
	}

	// Setup latest run manager
	latest, err := inmem.NewLatestRunManager(app, app)
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
	if err := mapper.Populate(app, app); err != nil {
		return nil, err
	}

	queues := inmem.NewWorkspaceQueueManager()
	if err := queues.Populate(app); err != nil {
		return nil, fmt.Errorf("populating workspace queues: %w", err)
	}
	app.queues = queues

	return app, nil
}
