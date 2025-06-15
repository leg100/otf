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
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/disco"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/forgejo"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/inmem"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		Config
		logr.Logger

		*sql.DB

		Organizations *organization.Service
		Runs          *run.Service
		Workspaces    *workspace.Service
		Variables     *variable.Service
		Notifications *notifications.Service
		Logs          *logs.Service
		State         *state.Service
		Configs       *configversion.Service
		Modules       *module.Service
		VCSProviders  *vcs.Service
		Tokens        *tokens.Service
		Teams         *team.Service
		Users         *user.Service
		GithubApp     *github.Service
		RepoHooks     *repohooks.Service
		Runners       *runner.Service
		Connections   *connections.Service
		System        *internal.HostnameService

		handlers []internal.Handlers
		listener *sql.Listener
		runner   *runner.Runner
	}
)

// New builds a new daemon and establishes a connection to the database and
// migrates it to the latest schema. Close() should be called to close this
// connection.
func New(ctx context.Context, logger logr.Logger, cfg Config) (*Daemon, error) {
	if internal.DevMode {
		logger.Info("enabled developer mode")
	}
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	logger.V(1).Info("set engine type", "engine", cfg.DefaultEngine)

	hostnameService := internal.NewHostnameService(cfg.Host)
	hostnameService.SetWebhookHostname(cfg.WebhookHost)

	cache, err := inmem.NewCache(*cfg.CacheConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("started cache", "max_size", cfg.CacheConfig.Size, "ttl", cfg.CacheConfig.TTL)

	db, err := sql.New(ctx, logger, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("creating database pool: %w", err)
	}

	// listener listens to database events
	listener := sql.NewListener(logger, db)

	// responder responds to TFE API requests
	responder := tfeapi.NewResponder()

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

	authorizer := authz.NewAuthorizer(logger)

	orgService := organization.NewService(organization.Options{
		Logger:                       logger,
		Authorizer:                   authorizer,
		DB:                           db,
		Listener:                     listener,
		Responder:                    responder,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
		TokensService:                tokensService,
	})

	teamService := team.NewService(team.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		Responder:           responder,
		OrganizationService: orgService,
		TokensService:       tokensService,
	})
	userService := user.NewService(user.Options{
		Logger:        logger,
		Authorizer:    authorizer,
		DB:            db,
		Responder:     responder,
		TokensService: tokensService,
		SiteToken:     cfg.SiteToken,
		TeamService:   teamService,
	})
	// promote nominated users to site admin
	if err := userService.SetSiteAdmins(ctx, cfg.SiteAdmins...); err != nil {
		return nil, err
	}

	configService := configversion.NewService(configversion.Options{
		Logger:        logger,
		Authorizer:    authorizer,
		DB:            db,
		Responder:     responder,
		Cache:         cache,
		Signer:        signer,
		MaxConfigSize: cfg.MaxConfigSize,
	})

	vcsService := vcs.NewService(vcs.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		Responder:           responder,
		HostnameService:     hostnameService,
		SourceIconRegistrar: configService,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})

	vcsEventBroker := &vcs.Broker{}

	githubAppService := github.NewService(github.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		HostnameService:     hostnameService,
		VCSService:          vcsService,
		GithubHostname:      cfg.GithubHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		VCSEventBroker:      vcsEventBroker,
	})

	repoService := repohooks.NewService(ctx, repohooks.Options{
		Logger:              logger,
		DB:                  db,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
		VCSService:          vcsService,
		GithubAppService:    githubAppService,
		VCSEventBroker:      vcsEventBroker,
	})

	connectionService := connections.NewService(ctx, connections.Options{
		Logger:             logger,
		DB:                 db,
		VCSProviderService: vcsService,
		RepoHooksService:   repoService,
	})
	engineService := engine.NewService(engine.Options{
		Logger: logger,
		DB:     db,
	})
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		engineService.StartLatestChecker(ctx)
	}
	workspaceService := workspace.NewService(workspace.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		Listener:            listener,
		Responder:           responder,
		ConnectionService:   connectionService,
		TeamService:         teamService,
		UserService:         userService,
		OrganizationService: orgService,
		VCSProviderService:  vcsService,
		DefaultEngine:       cfg.DefaultEngine,
		EngineService:       engineService,
	})

	logsService := logs.NewService(logs.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Cache:      cache,
		Listener:   listener,
		Verifier:   signer,
	})
	runService := run.NewService(run.Options{
		Logger:               logger,
		Authorizer:           authorizer,
		DB:                   db,
		Listener:             listener,
		Responder:            responder,
		OrganizationService:  orgService,
		WorkspaceService:     workspaceService,
		LogsService:          logsService,
		ConfigVersionService: configService,
		VCSProviderService:   vcsService,
		Cache:                cache,
		VCSEventSubscriber:   vcsEventBroker,
		Signer:               signer,
		EngineService:        engineService,
		TokensService:        tokensService,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		Authorizer:         authorizer,
		DB:                 db,
		HostnameService:    hostnameService,
		VCSProviderService: vcsService,
		Signer:             signer,
		ConnectionsService: connectionService,
		RepohookService:    repoService,
		VCSEventSubscriber: vcsEventBroker,
	})
	stateService := state.NewService(state.Options{
		Logger:           logger,
		Authorizer:       authorizer,
		DB:               db,
		WorkspaceService: workspaceService,
		Cache:            cache,
		Responder:        responder,
		Signer:           signer,
	})
	variableService := variable.NewService(variable.Options{
		Logger:           logger,
		Authorizer:       authorizer,
		DB:               db,
		Responder:        responder,
		WorkspaceService: workspaceService,
		RunClient:        runService,
	})
	runnerService := runner.NewService(runner.ServiceOptions{
		Logger:           logger,
		Authorizer:       authorizer,
		DB:               db,
		Responder:        responder,
		RunService:       runService,
		WorkspaceService: workspaceService,
		TokensService:    tokensService,
		Listener:         listener,
	})
	authenticatorService, err := authenticator.NewAuthenticatorService(ctx, authenticator.Options{
		Logger:               logger,
		HostnameService:      hostnameService,
		TokensService:        tokensService,
		UserService:          userService,
		IDTokenHandlerConfig: cfg.OIDC,
		SkipTLSVerification:  cfg.SkipTLSVerification,
	})
	if err != nil {
		return nil, err
	}

	// Forgejo registrations
	forgejo.RegisterVCSKind(
		vcsService,
		cfg.ForgejoHostname,
		cfg.SkipTLSVerification,
	)

	// Gitlab registrations
	gitlab.RegisterVCSKind(
		vcsService,
		cfg.GitlabHostname,
		cfg.SkipTLSVerification,
	)
	err = gitlab.RegisterOAuthHandler(
		authenticatorService,
		cfg.GitlabHostname,
		cfg.GitlabClientID,
		cfg.GitlabClientSecret,
		cfg.SkipTLSVerification,
	)
	if err != nil {
		return nil, fmt.Errorf("registering gitlab oauth client: %w", err)
	}

	// Github registrations
	err = github.RegisterOAuthHandler(
		authenticatorService,
		cfg.GithubHostname,
		cfg.GithubClientID,
		cfg.GithubClientSecret,
		cfg.SkipTLSVerification,
	)
	if err != nil {
		return nil, fmt.Errorf("registering github oauth client: %w", err)
	}

	runner, err := runner.NewServerRunner(runner.ServerRunnerOptions{
		Logger:     logger,
		Config:     cfg.RunnerConfig,
		Runners:    runnerService,
		Workspaces: workspaceService,
		Variables:  variableService,
		State:      stateService,
		Configs:    configService,
		Runs:       runService,
		Logs:       logsService,
		Jobs:       runnerService,
		Server:     hostnameService,
	})
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Listener:   listener,
		Responder:  responder,
	})

	handlers := []internal.Handlers{
		teamService,
		userService,
		workspaceService,
		stateService,
		orgService,
		variableService,
		vcsService,
		moduleService,
		runService,
		logsService,
		repoService,
		authenticatorService,
		loginserver.NewServer(loginserver.Options{
			Secret:      cfg.Secret,
			UserService: userService,
		}),
		configService,
		notificationService,
		githubAppService,
		runnerService,
		disco.Service{},
		&api.Handlers{},
		&tfeapi.Handlers{},
		&github.AppEventHandler{
			Logger:     logger,
			Publisher:  vcsEventBroker,
			GithubApps: githubAppService,
			VCSService: vcsService,
		},
	}

	return &Daemon{
		Config:        cfg,
		Logger:        logger,
		handlers:      handlers,
		Organizations: orgService,
		System:        hostnameService,
		Runs:          runService,
		Workspaces:    workspaceService,
		Variables:     variableService,
		Notifications: notificationService,
		Logs:          logsService,
		State:         stateService,
		Configs:       configService,
		Modules:       moduleService,
		VCSProviders:  vcsService,
		Tokens:        tokensService,
		Teams:         teamService,
		Users:         userService,
		RepoHooks:     repoService,
		GithubApp:     githubAppService,
		Connections:   connectionService,
		Runners:       runnerService,
		DB:            db,
		runner:        runner,
		listener:      listener,
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
		Middleware:           []mux.MiddlewareFunc{d.Tokens.Middleware()},
		Handlers:             d.handlers,
		AllowedOrigins:       d.AllowedOrigins,
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
		d.System.SetHostname(internal.NormalizeAddress(listenAddress))
	}

	d.V(0).Info("set system hostname", "hostname", d.System.Hostname())
	d.V(0).Info("set webhook hostname", "webhook_hostname", d.System.WebhookHostname())

	// Start subsystems. Subsystems are started in order.
	subsystems := []*Subsystem{
		// The listener is started first because it is responsible for listening
		// for database events, and other subsystems rely on it to be listening
		// before they start, such as the runner (which generates a new runner
		// event), and the allocator (which receives the new runner event).
		{
			Name:   "listener",
			Logger: d.Logger,
			System: d.listener,
		},
		{
			Name:   "reporter",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.ReporterLockID),
			System: &run.Reporter{
				Logger:          d.Logger.WithValues("component", "reporter"),
				VCS:             d.VCSProviders,
				HostnameService: d.System,
				Workspaces:      d.Workspaces,
				Runs:            d.Runs,
				Configs:         d.Configs,
				Cache:           make(map[resource.TfeID]vcs.Status),
			},
		},
		{
			Name:   "run_metrics",
			Logger: d.Logger,
			System: &run.MetricsCollector{
				Service: d.Runs,
			},
		},
		{
			Name:   "timeout",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.TimeoutLockID),
			System: &run.Timeout{
				Logger:                d.Logger.WithValues("component", "timeout"),
				OverrideCheckInterval: d.OverrideTimeoutCheckInterval,
				PlanningTimeout:       d.PlanningTimeout,
				ApplyingTimeout:       d.ApplyingTimeout,
				Runs:                  d.Runs,
			},
		},
		{
			Name:   "notifier",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.NotifierLockID),
			System: notifications.NewNotifier(notifications.NotifierOptions{
				Logger:             d.Logger,
				HostnameService:    d.System,
				WorkspaceClient:    d.Workspaces,
				RunClient:          d.Runs,
				NotificationClient: d.Notifications,
				DB:                 d.DB,
			}),
		},
		{
			Name:   "job-allocator",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.AllocatorLockID),
			System: d.Runners.NewAllocator(d.Logger),
		},
		{
			Name:   "runner-manager",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.RunnerManagerLockID),
			System: d.Runners.NewManager(),
		},
	}
	if !d.DisableRunner {
		subsystems = append(subsystems, &Subsystem{
			Name:   "runner-daemon",
			Logger: d.Logger,
			DB:     d.DB,
			System: d.runner,
		})
	}
	if !d.DisableScheduler {
		subsystems = append(subsystems, &Subsystem{
			Name:   "scheduler",
			Logger: d.Logger,
			DB:     d.DB,
			LockID: internal.Int64(sql.SchedulerLockID),
			System: run.NewScheduler(run.SchedulerOptions{
				Logger:          d.Logger,
				WorkspaceClient: d.Workspaces,
				RunClient:       d.Runs,
			}),
		})
	}
	for _, ss := range subsystems {
		if err := ss.Start(ctx, g); err != nil {
			return err
		}
		// Wait for subsystem to finish starting up if it exposes the ability to
		// do so.
		wait, ok := ss.System.(interface{ Started() <-chan struct{} })
		if ok {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * 10):
				return fmt.Errorf("timed out waiting for subsystem to start: %s", ss.Name)
			case <-wait.Started():
			}
		}
	}

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
