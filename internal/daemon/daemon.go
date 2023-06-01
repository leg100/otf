// Package daemon configures and starts the otfd daemon and its subsystems.
package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/disco"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/orgcreator"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/scheduler"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		Config
		logr.Logger

		internal.DB
		*pubsub.Broker

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
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		repo.RepoService
		logs.LogsService
		notifications.NotificationService

		Handlers []internal.Handlers
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
	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	hostnameService := internal.NewHostnameService(cfg.Host)

	renderer, err := html.NewRenderer(cfg.DevMode)
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
	signer := internal.NewSigner(cfg.Secret)

	orgService := organization.NewService(organization.Options{
		Logger:                       logger,
		DB:                           db,
		Renderer:                     renderer,
		Broker:                       broker,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
	})

	authService := auth.NewService(auth.Options{
		Logger:          logger,
		DB:              db,
		Renderer:        renderer,
		HostnameService: hostnameService,
	})
	// promote nominated users to site admin
	if err := authService.SetSiteAdmins(ctx, cfg.SiteAdmins...); err != nil {
		return nil, err
	}

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
		TokensService:              tokensService,
		Configs:                    []cloud.CloudOAuthConfig{cfg.Github, cfg.Gitlab},
		OIDCConfigs:                []cloud.OIDCConfig{cfg.OIDC},
	})
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:              logger,
		DB:                  db,
		Broker:              broker,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		HostnameService:     hostnameService,
	})

	loginServer, err := loginserver.NewServer(loginserver.Options{
		Secret:        cfg.Secret,
		Renderer:      renderer,
		TokensService: tokensService,
	})
	if err != nil {
		return nil, err
	}

	api := api.New(api.Options{
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		OrganizationCreatorService:  orgCreatorService,
		StateService:                stateService,
		RunService:                  runService,
		ConfigurationVersionService: configService,
		AuthService:                 authService,
		TokensService:               tokensService,
		VariableService:             variableService,
		NotificationService:         notificationService,
		Signer:                      signer,
		MaxConfigSize:               cfg.MaxConfigSize,
	})

	handlers := []internal.Handlers{
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
		moduleService,
		runService,
		logsService,
		repoService,
		authenticatorService,
		loginServer,
		disco.Service{},
		api,
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
		NotificationService:         notificationService,
		Broker:                      broker,
		DB:                          db,
		agent:                       agent,
	}, nil
}

// Start the otfd daemon and block until ctx is cancelled or an error is
// returned. The started channel is closed once the daemon has started.
func (d *Daemon) Start(ctx context.Context, started chan struct{}) error {
	// Cancel context the first time a func started with g.Go() fails
	g, ctx := errgroup.WithContext(ctx)

	// close all db connections upon exit
	defer d.DB.Close()

	// Construct web server and start listening on port
	server, err := http.NewServer(d.Logger, http.ServerConfig{
		SSL:                  d.SSL,
		CertFile:             d.CertFile,
		KeyFile:              d.KeyFile,
		EnableRequestLogging: d.EnableRequestLogging,
		DevMode:              d.DevMode,
		Middleware:           []mux.MiddlewareFunc{d.TokensService.Middleware()},
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
		d.SetHostname(internal.NormalizeAddress(listenAddress))
	}
	d.V(0).Info("set system hostname", "hostname", d.Hostname())

	subsystems := []*Subsystem{
		{
			Name:           "broker",
			BackoffRestart: true,
			Logger:         d.Logger,
			System:         d.Broker,
		},
		{
			Name:           "proxy",
			BackoffRestart: true,
			Logger:         d.Logger,
			System:         d.LogsService,
		},
		{
			Name:           "spawner",
			BackoffRestart: true,
			Logger:         d.Logger,
			System: &run.Spawner{
				Logger:                      d.Logger.WithValues("component", "spawner"),
				ConfigurationVersionService: d.ConfigurationVersionService,
				WorkspaceService:            d.WorkspaceService,
				VCSProviderService:          d.VCSProviderService,
				RunService:                  d.RunService,
				Subscriber:                  d.Broker,
			},
		},
		{
			Name:           "publisher",
			BackoffRestart: true,
			Logger:         d.Logger,
			System: &module.Publisher{
				Logger:             d.Logger.WithValues("component", "publisher"),
				VCSProviderService: d.VCSProviderService,
				ModuleService:      d.ModuleService,
				Subscriber:         d.Broker,
			},
		},
		{
			Name:           "reporter",
			BackoffRestart: true,
			Logger:         d.Logger,
			Exclusive:      true,
			DB:             d.DB,
			LockID:         internal.Int64(run.ReporterLockID),
			System: &run.Reporter{
				Logger:                      d.Logger.WithValues("component", "reporter"),
				VCSProviderService:          d.VCSProviderService,
				Subscriber:                  d.Broker,
				HostnameService:             d.HostnameService,
				ConfigurationVersionService: d.ConfigurationVersionService,
				WorkspaceService:            d.WorkspaceService,
			},
		},
		{
			Name:           "notifier",
			BackoffRestart: true,
			Logger:         d.Logger,
			Exclusive:      true,
			DB:             d.DB,
			LockID:         internal.Int64(notifications.LockID),
			System: notifications.NewNotifier(notifications.NotifierOptions{
				Logger:           d.Logger,
				Subscriber:       d.Broker,
				HostnameService:  d.HostnameService,
				WorkspaceService: d.WorkspaceService,
				DB:               d.DB,
			}),
		},
	}
	if !d.DisableScheduler {
		subsystems = append(subsystems, &Subsystem{
			Name:           "scheduler",
			BackoffRestart: true,
			Logger:         d.Logger,
			Exclusive:      true,
			DB:             d.DB,
			LockID:         internal.Int64(scheduler.LockID),
			System: scheduler.NewScheduler(scheduler.Options{
				Logger:           d.Logger,
				WorkspaceService: d.WorkspaceService,
				RunService:       d.RunService,
				DB:               d.DB,
				Subscriber:       d,
			}),
		})
	}
	for _, ss := range subsystems {
		if err := ss.Start(ctx, g); err != nil {
			return err
		}
	}

	// Wait for broker start listening; otherwise some tests may fail
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 10):
		return fmt.Errorf("timed out waiting for broker to start")
	case <-d.Broker.Started():
	}

	// Run local agent in background
	g.Go(func() error {
		// give local agent unlimited access to services
		agentCtx := internal.AddSubjectToContext(ctx, &internal.Superuser{Username: "local-agent"})
		if err := d.agent.Start(agentCtx); err != nil {
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
