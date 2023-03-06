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
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/repo"
)

var _ otf.Application = (*Application)(nil)

// Application encompasses services for interacting between components of the
// otf server
type Application struct {
	db       otf.DB
	cache    otf.Cache
	proxy    otf.ChunkStore
	hostname string

	opts Options // keep reference for creating child apps

	*otf.RunFactory
	*otf.VCSProviderFactory
	otf.RepoService
	*otf.RunStarter
	*module.Publisher
	*otf.ModuleVersionUploader
	otf.ModuleDeleter
	cloud.Service
	otf.PubSubService
	logr.Logger
	otf.Authorizer
	otf.StateVersionService
	otf.VariableApp
}

// NewApplication constructs an application, initialising various services and
// daemons.
func NewApplication(ctx context.Context, opts Options) (*Application, error) {
	// Any services that use transactions or advisory locks should be
	// constructed via newApp
	app := newChildApp(&Application{}, opts, opts.DB)

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
func newChildApp(parent *Application, opts Options, db otf.DB) *Application {
	child := &Application{
		Logger:              opts.Logger,
		cache:               opts.Cache,
		PubSubService:       opts.PubSub,
		Service:             opts.CloudService,
		Authorizer:          otf.NewAuthorizer(opts.Logger, db),
		StateVersionService: opts.StateVersionService,
		RunFactory:          parent.RunFactory,
		proxy:               parent.proxy,
		hostname:            parent.hostname,
		VCSProviderFactory: &otf.VCSProviderFactory{
			Service: opts.CloudService,
		},
		db:   db,
		opts: opts,
	}
	child.RunFactory = &otf.RunFactory{
		WorkspaceService:            child,
		ConfigurationVersionService: child,
	}
	child.RepoService = repo.NewService(repo.NewServiceOptions{
		Logger:             opts.Logger,
		Database:           db,
		CloudService:       child.Service,
		HostnameService:    child,
		VCSProviderService: child,
	})
	child.Publisher = module.NewPublisher(child)
	child.RunStarter = &otf.RunStarter{child}
	child.ModuleVersionUploader = &otf.ModuleVersionUploader{child}
	child.ModuleDeleter = module.NewDeleter(child)

	return child
}

type Options struct {
	Logger              logr.Logger
	DB                  otf.DB
	Cache               *bigcache.BigCache
	PubSub              otf.PubSubService
	CloudService        cloud.Service
	Authorizer          otf.Authorizer
	StateVersionService otf.StateVersionService
}

func (a *Application) DB() otf.DB { return a.db }

// Tx provides a callback in which all db interactions are wrapped within a
// transaction. Useful for ensuring multiple service calls succeed together.
func (a *Application) Tx(ctx context.Context, tx func(a otf.Application) error) error {
	return a.db.Tx(ctx, func(db otf.DB) error {
		// wrap copy of app inside tx
		return tx(newChildApp(a, a.opts, db))
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
		return cb(newChildApp(a, a.opts, db))
	})
}
