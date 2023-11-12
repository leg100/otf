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
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/disco"
	"github.com/leg100/otf/internal/ghapphandler"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/remoteops"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/scheduler"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		Config
		logr.Logger

		*sql.DB
		*pubsub.Broker

		organization.OrganizationService
		team.TeamService
		user.UserService
		tokens.TokensService
		variable.VariableService
		vcsprovider.VCSProviderService
		state.StateService
		workspace.WorkspaceService
		module.ModuleService
		internal.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		repohooks.RepohookService
		logs.LogsService
		notifications.NotificationService
		connections.ConnectionService
		github.GithubAppService
		remoteops.AgentTokenService

		Handlers []internal.Handlers

		opsDaemon process
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

	responder := tfeapi.NewResponder()

	broker := pubsub.NewBroker(logger, db)

	// Setup url signer
	signer := internal.NewSigner(cfg.Secret)

	tokensService, err := tokens.NewService(tokens.Options{
		Logger:          logger,
		GoogleIAPConfig: cfg.GoogleIAPConfig,
		Secret:          cfg.Secret,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up authentication middleware: %w", err)
	}

	orgService := organization.NewService(organization.Options{
		Logger:                       logger,
		DB:                           db,
		Renderer:                     renderer,
		Responder:                    responder,
		Broker:                       broker,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
		TokensService:                tokensService,
	})

	teamService := team.NewService(team.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Responder:           responder,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
		TokensService:       tokensService,
	})
	userService := user.NewService(user.Options{
		Logger:          logger,
		DB:              db,
		Renderer:        renderer,
		Responder:       responder,
		HostnameService: hostnameService,
		TokensService:   tokensService,
		SiteToken:       cfg.SiteToken,
		TeamService:     teamService,
	})
	// promote nominated users to site admin
	if err := userService.SetSiteAdmins(ctx, cfg.SiteAdmins...); err != nil {
		return nil, err
	}

	githubAppService := github.NewService(github.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		HostnameService:     hostnameService,
		GithubHostname:      cfg.GithubHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})

	vcsEventBroker := &vcs.Broker{}

	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Responder:           responder,
		HostnameService:     hostnameService,
		GithubAppService:    githubAppService,
		GithubHostname:      cfg.GithubHostname,
		GitlabHostname:      cfg.GitlabHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		Subscriber:          vcsEventBroker,
	})
	repoService := repohooks.NewService(ctx, repohooks.Options{
		Logger:              logger,
		DB:                  db,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
		GithubAppService:    githubAppService,
		VCSEventBroker:      vcsEventBroker,
	})
	repoService.RegisterCloudHandler(vcs.GithubKind, github.HandleEvent)
	repoService.RegisterCloudHandler(vcs.GitlabKind, gitlab.HandleEvent)

	connectionService := connections.NewService(ctx, connections.Options{
		Logger:             logger,
		DB:                 db,
		VCSProviderService: vcsProviderService,
		RepohookService:    repoService,
	})
	releasesService := releases.NewService(releases.Options{
		Logger: logger,
		DB:     db,
	})
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		releasesService.StartLatestChecker(ctx)
	}
	workspaceService := workspace.NewService(workspace.Options{
		Logger:              logger,
		DB:                  db,
		Broker:              broker,
		Renderer:            renderer,
		Responder:           responder,
		ConnectionService:   connectionService,
		TeamService:         teamService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
	})
	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Responder:           responder,
		Cache:               cache,
		Signer:              signer,
		MaxConfigSize:       cfg.MaxConfigSize,
	})

	runService := run.NewService(run.Options{
		Logger:                      logger,
		DB:                          db,
		Renderer:                    renderer,
		Responder:                   responder,
		WorkspaceAuthorizer:         workspaceService,
		OrganizationService:         orgService,
		WorkspaceService:            workspaceService,
		ConfigurationVersionService: configService,
		VCSProviderService:          vcsProviderService,
		Broker:                      broker,
		Cache:                       cache,
		VCSEventSubscriber:          vcsEventBroker,
		Signer:                      signer,
		ReleasesService:             releasesService,
		TokensService:               tokensService,
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
		ConnectionService:  connectionService,
		RepohookService:    repoService,
		VCSEventSubscriber: vcsEventBroker,
	})
	stateService := state.NewService(state.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		Cache:               cache,
		Renderer:            renderer,
		Responder:           responder,
		Signer:              signer,
	})
	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Responder:           responder,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		RunService:          runService,
	})
	agentTokenService := remoteops.NewService(remoteops.ServiceOptions{
		Logger:        logger,
		DB:            db,
		Renderer:      renderer,
		Responder:     responder,
		TokensService: tokensService,
	})

	remoteopsDaemon, err := remoteops.NewDaemon(
		logger.WithValues("component", "remoteops"),
		remoteops.InProcClient{
			WorkspaceService:            workspaceService,
			VariableService:             variableService,
			StateService:                stateService,
			HostnameService:             hostnameService,
			ConfigurationVersionService: configService,
			RunService:                  runService,
			LogsService:                 logsService,
		},
		*cfg.RemoteOpsConfig,
	)
	if err != nil {
		return nil, err
	}

	authenticatorService, err := authenticator.NewAuthenticatorService(ctx, authenticator.Options{
		Logger:          logger,
		Renderer:        renderer,
		HostnameService: hostnameService,
		TokensService:   tokensService,
		OpaqueHandlerConfigs: []authenticator.OpaqueHandlerConfig{
			{
				ClientConstructor: github.NewOAuthClient,
				OAuthConfig: authenticator.OAuthConfig{
					Hostname:     cfg.GithubHostname,
					Name:         string(vcs.GithubKind),
					Endpoint:     github.OAuthEndpoint,
					Scopes:       github.OAuthScopes,
					ClientID:     cfg.GithubClientID,
					ClientSecret: cfg.GithubClientSecret,
				},
			},
			{
				ClientConstructor: gitlab.NewOAuthClient,
				OAuthConfig: authenticator.OAuthConfig{
					Hostname:     cfg.GitlabHostname,
					Name:         string(vcs.GitlabKind),
					Endpoint:     gitlab.OAuthEndpoint,
					Scopes:       gitlab.OAuthScopes,
					ClientID:     cfg.GitlabClientID,
					ClientSecret: cfg.GitlabClientSecret,
				},
			},
		},
		IDTokenHandlerConfig: cfg.OIDC,
		SkipTLSVerification:  cfg.SkipTLSVerification,
	})
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:              logger,
		DB:                  db,
		Broker:              broker,
		Responder:           responder,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		HostnameService:     hostnameService,
	})

	loginServer, err := loginserver.NewServer(loginserver.Options{
		Secret:      cfg.Secret,
		Renderer:    renderer,
		UserService: userService,
	})
	if err != nil {
		return nil, err
	}

	handlers := []internal.Handlers{
		teamService,
		userService,
		workspaceService,
		stateService,
		orgService,
		variableService,
		vcsProviderService,
		moduleService,
		runService,
		logsService,
		repoService,
		authenticatorService,
		loginServer,
		configService,
		notificationService,
		githubAppService,
		agentTokenService,
		disco.Service{},
		&ghapphandler.Handler{
			Logger:             logger,
			Publisher:          vcsEventBroker,
			GithubAppService:   githubAppService,
			VCSProviderService: vcsProviderService,
		},
		&api.Handlers{},
		&tfeapi.Handlers{},
	}

	return &Daemon{
		Config:                      cfg,
		Logger:                      logger,
		Handlers:                    handlers,
		TeamService:                 teamService,
		UserService:                 userService,
		TokensService:               tokensService,
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		VariableService:             variableService,
		VCSProviderService:          vcsProviderService,
		StateService:                stateService,
		ModuleService:               moduleService,
		HostnameService:             hostnameService,
		ConfigurationVersionService: configService,
		RunService:                  runService,
		LogsService:                 logsService,
		RepohookService:             repoService,
		NotificationService:         notificationService,
		GithubAppService:            githubAppService,
		ConnectionService:           connectionService,
		AgentTokenService:           agentTokenService,
		Broker:                      broker,
		DB:                          db,
		opsDaemon:                   remoteopsDaemon,
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
			Name:   "broker",
			Logger: d.Logger,
			System: d.Broker,
		},
		{
			Name:   "proxy",
			Logger: d.Logger,
			System: d.LogsService,
		},
		{
			Name:      "reporter",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(run.ReporterLockID),
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
			Name:      "notifier",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(notifications.LockID),
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
			Name:      "scheduler",
			Logger:    d.Logger,
			Exclusive: true,
			DB:        d.DB,
			LockID:    internal.Int64(scheduler.LockID),
			System: scheduler.NewScheduler(scheduler.Options{
				Logger:           d.Logger,
				WorkspaceService: d.WorkspaceService,
				RunService:       d.RunService,
				Subscriber:       d.Broker,
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

	// Run remote ops daemon in background
	g.Go(func() error {
		// give daemon unlimited access to services
		daemonCtx := internal.AddSubjectToContext(ctx, &internal.Superuser{Username: "remoteops-daemon"})
		if err := d.opsDaemon.Start(daemonCtx); err != nil {
			return fmt.Errorf("remote ops daemon terminated: %w", err)
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
