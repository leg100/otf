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
	"github.com/leg100/otf/hooks"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/workspace"
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
	otf.WorkspaceConnector
	otf.HookService
	*otf.RunStarter
	*module.Publisher
	*otf.ModuleVersionUploader
	otf.ModuleDeleter
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
		Authorizer:    &authorizer{opts.DB, opts.Logger},
	}
	// Any services that use transactions or advisory locks should be
	// constructed via newApp
	app = newChildApp(app, opts.DB)

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

// newChildApp constructs a child app with a specific db connection, e.g. a
// transaction or one holding an advisory lock, assigning the connection to the
// constituent services that comprise an app for the services to make use of
// that connection. May be called multiple times e.g. for nesting transactions.
func newChildApp(parent *Application, db otf.DB) *Application {
	child := &Application{
		PubSubService: parent.PubSubService,
		cache:         parent.cache,
		Logger:        parent.Logger,
		RunFactory:    parent.RunFactory,
		Authorizer:    parent.Authorizer,
		proxy:         parent.proxy,
		hostname:      parent.hostname,
		Service:       parent.Service,
		VCSProviderFactory: &otf.VCSProviderFactory{
			Service: parent.Service,
		},
		db: db,
	}
	child.RunFactory = &otf.RunFactory{
		WorkspaceService:            child,
		ConfigurationVersionService: child,
	}
	child.HookService = hooks.NewService(hooks.NewServiceOptions{
		Database:        db,
		CloudService:    child.Service,
		HostnameService: child,
	})
	child.WorkspaceConnector = &workspace.Connector{
		HookService:        child,
		WorkspaceService:   child,
		VCSProviderService: child,
	}
	child.Publisher = module.NewPublisher(child)
	child.RunStarter = &otf.RunStarter{child}
	child.ModuleVersionUploader = &otf.ModuleVersionUploader{child}
	child.ModuleDeleter = module.NewDeleter(child)

	return child
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
		// wrap copy of app inside tx
		return tx(newChildApp(a, db))
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
		return cb(newChildApp(a, db))
	})
}
