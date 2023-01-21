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
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/inmem"
)

var _ otf.Application = (*Application)(nil)

// Application encompasses services for interacting between components of the
// otf server
type Application struct {
	db       otf.DB
	cache    otf.Cache
	proxy    otf.ChunkStore
	hostname string

	*otf.RunFactory
	*otf.VCSProviderFactory
	*otf.WorkspaceConnector
	*otf.RunStarter
	*otf.Publisher
	*otf.ModuleVersionUploader
	*otf.WebhookSynchroniser
	cloud.Service
	otf.PubSubService
	logr.Logger
	Authorizer
}

// NewApplication constructs an application, initialising various services and
// daemons.
func NewApplication(ctx context.Context, opts Options) (*Application, error) {
	app := &Application{
		PubSubService: opts.PubSub,
		cache:         opts.Cache,
		db:            opts.DB,
		Logger:        opts.Logger,
		Service:       opts.CloudService,
	}
	app.Authorizer = &authorizer{opts.DB, opts.Logger}
	app.RunFactory = &otf.RunFactory{
		WorkspaceService:            app,
		ConfigurationVersionService: app,
	}
	app.VCSProviderFactory = &otf.VCSProviderFactory{
		Service: opts.CloudService,
	}
	app.WorkspaceConnector = &otf.WorkspaceConnector{
		Application: app,
		WebhookCreator: &otf.WebhookCreator{
			VCSProviderService: app,
			Service:            opts.CloudService,
			HostnameService:    app,
		},
		WebhookUpdater: &otf.WebhookUpdater{
			VCSProviderService: app,
			HostnameService:    app,
		},
	}
	app.Publisher = otf.NewPublisher(app)
	app.WebhookSynchroniser = &otf.WebhookSynchroniser{app}
	app.RunStarter = &otf.RunStarter{app}
	app.ModuleVersionUploader = &otf.ModuleVersionUploader{app}

	proxy, err := inmem.NewChunkProxy(app, opts.Logger, opts.Cache, opts.DB)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	app.proxy = proxy
	go func() {
		if err := proxy.Start(ctx); ctx.Err() == nil {
			app.Error(err, "proxy unexpectedly terminated")
		}
	}()

	return app, nil
}

type Options struct {
	Logger       logr.Logger
	DB           otf.DB
	Cache        *bigcache.BigCache
	PubSub       otf.PubSubService
	CloudService cloud.Service
}

func (a *Application) DB() otf.DB { return a.db }

// Tx provides a callback in which all db interactions are wrapped within a
// transaction. Useful for ensuring multiple service calls succeed together.
func (a *Application) Tx(ctx context.Context, tx func(a otf.Application) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		// make a copy of the app and assign a db tx wrapper
		appTx := &Application{
			PubSubService: a.PubSubService,
			cache:         a.cache,
			Logger:        a.Logger,
			RunFactory:    a.RunFactory,
			Authorizer:    a.Authorizer,
			proxy:         a.proxy,
			db:            db,
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
			PubSubService: a.PubSubService,
			cache:         a.cache,
			Logger:        a.Logger,
			RunFactory:    a.RunFactory,
			Authorizer:    a.Authorizer,
			proxy:         a.proxy,
			db:            db,
		}
		return cb(appWithLock)
	})
}
