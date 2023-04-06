// Package daemon configures and starts the otfd daemon and its subsystems.
package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/authenticator"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/tokens"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		Config
		logr.Logger

		otf.DB
		*pubsub.Broker
		authenticator.Synchroniser

		agent process

		organization.OrganizationService
		orgcreator.OrganizationCreatorService
		auth.AuthService
		tokens.TokensService
		variable.VariableService
		vcsprovider.VCSProviderService
		state.StateService
		workspace.WorkspaceService
		module.ModuleService
		otf.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		repo.RepoService
		logs.LogsService

		Handlers []otf.Handlers

		// AuthMiddleware protects authenticated routes
		AuthMiddleware mux.MiddlewareFunc
	}

	process interface {
		Start(context.Context) error
	}
)

// New builds a new daemon and establishes a connection to the database and
// migrates it to the latest schema. Close() should be called to close this
// connection.
func New(ctx context.Context, logger logr.Logger, cfg Config) (*Daemon, error) {
	if cfg.DevMode {
		logger.Info("enabled developer mode")
	}

	hostnameService := otf.NewHostnameService(cfg.Host)

	renderer, err := html.NewViewEngine(cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("setting up web page renderer: %w", err)
	}
	cloudService, err := inmem.NewCloudService(cfg.Github.Config, cfg.Gitlab.Config)
	if err != nil {
		return nil, err
	}
	cache, err := inmem.NewCache(*cfg.CacheConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("started cache", "max_size", cfg.CacheConfig.Size, "ttl", cfg.CacheConfig.TTL)

	db, err := sql.New(ctx, sql.Options{
		Logger:     logger,
		ConnString: cfg.Database,
	})
	if err != nil {
		return nil, err
	}

	broker := pubsub.NewBroker(logger, db)

	// Setup url signer
	signer := otf.NewSigner(cfg.Secret)

	orgService := organization.NewService(organization.Options{
		Logger:   logger,
		DB:       db,
		Renderer: renderer,
		Broker:   broker,
	})
	authService := auth.NewService(auth.Options{
		Logger:          logger,
		DB:              db,
		Renderer:        renderer,
		HostnameService: hostnameService,
	})

	tokensService, err := tokens.NewService(tokens.Options{
		Logger:          logger,
		DB:              db,
		Renderer:        renderer,
		AuthService:     authService,
		GoogleIAPConfig: cfg.GoogleIAPConfig,
		SiteToken:       cfg.SiteToken,
		Secret:          cfg.Secret,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up authentication middleware: %w", err)
	}

	orgCreatorService := orgcreator.NewService(orgcreator.Options{
		Logger:                       logger,
		DB:                           db,
		Renderer:                     renderer,
		Publisher:                    broker,
		AuthService:                  authService,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
	})
	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:       logger,
		DB:           db,
		Renderer:     renderer,
		CloudService: cloudService,
	})
	repoService := repo.NewService(repo.Options{
		Logger:             logger,
		DB:                 db,
		CloudService:       cloudService,
		HostnameService:    hostnameService,
		Publisher:          broker,
		VCSProviderService: vcsProviderService,
	})

	workspaceService := workspace.NewService(workspace.Options{
		Logger:              logger,
		DB:                  db,
		Broker:              broker,
		Renderer:            renderer,
		RepoService:         repoService,
		TeamService:         authService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
	})
	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
		Signer:              signer,
		MaxUploadSize:       cfg.MaxConfigSize,
	})
	runService := run.NewService(run.Options{
		Logger:                      logger,
		DB:                          db,
		Renderer:                    renderer,
		WorkspaceAuthorizer:         workspaceService,
		WorkspaceService:            workspaceService,
		ConfigurationVersionService: configService,
		VCSProviderService:          vcsProviderService,
		Broker:                      broker,
		Cache:                       cache,
		Signer:                      signer,
	})
	logsService := logs.NewService(logs.Options{
		Logger:        logger,
		DB:            db,
		RunAuthorizer: runService,
		Cache:         cache,
		Broker:        broker,
		Verifier:      signer,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		DB:                 db,
		Renderer:           renderer,
		HostnameService:    hostnameService,
		VCSProviderService: vcsProviderService,
		Signer:             signer,
		RepoService:        repoService,
	})
	stateService := state.NewService(state.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		Cache:               cache,
	})
	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
	})

	agent, err := agent.NewAgent(
		logger.WithValues("component", "agent"),
		client.LocalClient{
			AuthService:                 authService,
			TokensService:               tokensService,
			WorkspaceService:            workspaceService,
			OrganizationService:         orgService,
			VariableService:             variableService,
			StateService:                stateService,
			HostnameService:             hostnameService,
			ConfigurationVersionService: configService,
			RunService:                  runService,
			LogsService:                 logsService,
		},
		*cfg.AgentConfig,
	)
	if err != nil {
		return nil, err
	}

	authenticatorService, err := authenticator.NewAuthenticatorService(authenticator.Options{
		Logger:                     logger,
		Renderer:                   renderer,
		HostnameService:            hostnameService,
		OrganizationService:        orgService,
		OrganizationCreatorService: orgCreatorService,
		AuthService:                authService,
		Configs:                    []cloud.CloudOAuthConfig{cfg.Github, cfg.Gitlab},
	})
	if err != nil {
		return nil, err
	}

	handlers := []otf.Handlers{
		authService,
		tokensService,
		workspaceService,
		// deliberating placing org creator service prior to org service because
		// org creator adds web routes that take priority (gorilla mux routes
		// are checked in the order they are added to the router).
		orgCreatorService,
		orgService,
		variableService,
		vcsProviderService,
		stateService,
		moduleService,
		configService,
		runService,
		logsService,
		repoService,
		authenticatorService,
	}

	return &Daemon{
		Config:                      cfg,
		Logger:                      logger,
		Handlers:                    handlers,
		AuthService:                 authService,
		TokensService:               tokensService,
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		OrganizationCreatorService:  orgCreatorService,
		VariableService:             variableService,
		VCSProviderService:          vcsProviderService,
		StateService:                stateService,
		ModuleService:               moduleService,
		HostnameService:             hostnameService,
		ConfigurationVersionService: configService,
		RunService:                  runService,
		LogsService:                 logsService,
		RepoService:                 repoService,
		Synchroniser:                authenticatorService,
		Broker:                      broker,
		DB:                          db,
		agent:                       agent,
	}, nil
}

// Start the otfd daemon and block until ctx is cancelled or an error is
// returned. The started channel is closed once the daemon has started.
func (d *Daemon) Start(ctx context.Context, started chan struct{}) error {
	// Give superuser privileges to all server processes
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "app-user"})
	// Cancel context the first time a func started with g.Go() fails
	g, ctx := errgroup.WithContext(ctx)

	// close all db connections upon exit
	defer func() {
		d.DB.Close()
	}()

	// Construct web server and start listening on port
	server, err := http.NewServer(d.Logger, http.ServerConfig{
		SSL:                  d.SSL,
		CertFile:             d.CertFile,
		KeyFile:              d.KeyFile,
		EnableRequestLogging: d.EnableRequestLogging,
		DevMode:              d.DevMode,
		Middleware:           []mux.MiddlewareFunc{d.AuthMiddleware},
		Handlers:             d.Handlers,
	})
	if err != nil {
		return fmt.Errorf("setting up http server: %w", err)
	}
	ln, err := net.Listen("tcp", d.Address)
	if err != nil {
		return err
	}
	defer ln.Close()

	// Unless user has set a hostname, set the hostname to the listening address
	// of the http server.
	if d.Host == "" {
		listenAddress := ln.Addr().(*net.TCPAddr)
		d.SetHostname(otf.NormalizeAddress(listenAddress))
	}
	d.V(0).Info("set system hostname", "hostname", d.Hostname())

	// Run pubsub broker and wait for it to start listening
	isListening := make(chan struct{})
	g.Go(func() error { return d.Broker.Start(ctx, isListening) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return fmt.Errorf("timed out waiting for broker to start")
	case <-isListening:
	}

	// Run logs caching proxy
	g.Go(func() error { return d.StartProxy(ctx) })

	if !d.DisableRunScheduler {
		// Run scheduler - if there is another scheduler running already then
		// this'll wait until the other scheduler exits.
		g.Go(func() error {
			err := scheduler.Start(ctx, scheduler.Options{
				Logger:           d.Logger,
				WorkspaceService: d.WorkspaceService,
				RunService:       d.RunService,
				DB:               d.DB,
				Subscriber:       d,
			})
			if err != nil {
				return fmt.Errorf("scheduler terminated: %w", err)
			}
			d.V(2).Info("scheduler gracefully shutdown")
			return nil
		})
	}

	// Run run-spawner
	g.Go(func() error {
		err := run.StartSpawner(ctx, run.SpawnerOptions{
			Logger:                      d.Logger,
			ConfigurationVersionService: d.ConfigurationVersionService,
			WorkspaceService:            d.WorkspaceService,
			VCSProviderService:          d.VCSProviderService,
			RunService:                  d.RunService,
			Subscriber:                  d.Broker,
		})
		if err != nil {
			return fmt.Errorf("spawner terminated: %w", err)
		}
		d.V(2).Info("spawner gracefully shutdown")
		return nil
	})

	// Run module publisher
	g.Go(func() error {
		err := module.StartPublisher(ctx, module.PublisherOptions{
			Logger:             d.Logger,
			VCSProviderService: d.VCSProviderService,
			ModuleService:      d.ModuleService,
			Subscriber:         d.Broker,
		})
		if err != nil {
			return fmt.Errorf("module publisher terminated: %w", err)
		}
		d.V(2).Info("publisher gracefully shutdown")
		return nil
	})

	// Run PR reporter - if there is another reporter running already then
	// this'll wait until the other reporter exits.
	g.Go(func() error {
		err := run.StartReporter(ctx, run.ReporterOptions{
			Logger:                      d.Logger,
			ConfigurationVersionService: d.ConfigurationVersionService,
			WorkspaceService:            d.WorkspaceService,
			VCSProviderService:          d.VCSProviderService,
			HostnameService:             d.HostnameService,
			Subscriber:                  d.Broker,
			DB:                          d.DB,
		})
		if err != nil {
			return fmt.Errorf("reporter terminated: %w", err)
		}
		d.V(2).Info("reporter gracefully shutdown")
		return nil
	})

	// Run local agent in background
	g.Go(func() error {
		// give local agent unlimited access to services
		ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "local-agent"})

		if err := d.agent.Start(ctx); err != nil {
			return fmt.Errorf("agent terminated: %w", err)
		}
		return nil
	})

	// Run HTTP/JSON-API server and web app
	g.Go(func() error {
		if err := server.Start(ctx, ln); err != nil {
			return fmt.Errorf("http server terminated: %w", err)
		}
		return nil
	})

	// Inform the caller the daemon has started
	close(started)

	// Block until error or Ctrl-C received.
	return g.Wait()
}
