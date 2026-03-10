// Package daemon configures and starts the otfd daemon and its subsystems.
package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/disco"
	"github.com/leg100/otf/internal/dynamiccreds"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/forgejo"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sshkey"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/ui"
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
		SSHKeys       *sshkey.Service

		handlers    []internal.Handlers
		sqlListener *sql.Listener
		netListener net.Listener
		runner      *runner.Runner
		server      *http.Server
		subsystems  []*Subsystem
	}
)

// New builds a new daemon, establishes a connection to the database and
// migrates it to the latest schema, constructs services and subsystems and
// returns the daemon.
//
// Start() is then expected to be called to start the daemon.
//
// Close() must be called to release resources.
func New(ctx context.Context, logger logr.Logger, cfg Config) (*Daemon, error) {
	if internal.DevMode {
		logger.Info("enabled developer mode")
	}
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	logger.V(1).Info("set default engine", "engine", cfg.DefaultEngine)

	db, err := sql.New(ctx, logger, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("creating database pool: %w", err)
	}

	// sqlListener listens to database events.
	sqlListener := sql.NewListener(logger, db)

	// netListener opens a TCP port for listening on.
	netListener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	hostnameService := internal.NewHostnameService(
		logger,
		cfg.Host,
		cfg.WebhookHost,
		netListener.Addr().(*net.TCPAddr),
	)

	// responder responds to TFE API requests
	responder := tfeapi.NewResponder(logger)

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
		Listener:                     sqlListener,
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
		Signer:        signer,
		MaxConfigSize: cfg.MaxConfigSize,
	})

	vcsService := vcs.NewService(vcs.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		Responder:           responder,
		SourceIconRegistrar: configService,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})

	vcsEventBroker := &vcs.Broker{}

	githubAppService := github.NewService(github.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		VCSService:          vcsService,
		GithubAPIURL:        cfg.GithubHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		VCSEventBroker:      vcsEventBroker,
	})

	repoService := repohooks.NewService(ctx, repohooks.Options{
		Logger:              logger,
		DB:                  db,
		URLs:                hostnameService,
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
		Listener:            sqlListener,
		Responder:           responder,
		ConnectionService:   connectionService,
		TeamService:         teamService,
		UserService:         userService,
		OrganizationService: orgService,
		VCSProviderService:  vcsService,
		DefaultEngine:       cfg.DefaultEngine,
		EngineService:       engineService,
	})

	runService := run.NewService(run.Options{
		Logger:               logger,
		Authorizer:           authorizer,
		DB:                   db,
		Listener:             sqlListener,
		Responder:            responder,
		OrganizationService:  orgService,
		WorkspaceService:     workspaceService,
		ConfigVersionService: configService,
		VCSProviderService:   vcsService,
		VCSEventSubscriber:   vcsEventBroker,
		Signer:               signer,
		EngineService:        engineService,
		TokensService:        tokensService,
		UsersService:         userService,
		DaemonCtx:            ctx,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		Authorizer:         authorizer,
		DB:                 db,
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
	dynamiccredsService, err := dynamiccreds.NewService(dynamiccreds.Options{
		HostnameService: hostnameService,
		PublicKeyPath:   cfg.PublicKeyPath,
		PrivateKeyPath:  cfg.PrivateKeyPath,
		Logger:          logger,
	})
	if err != nil {
		return nil, err
	}
	runnerService := runner.NewService(runner.ServiceOptions{
		Logger:                    logger,
		Authorizer:                authorizer,
		DB:                        db,
		Responder:                 responder,
		RunService:                runService,
		WorkspaceService:          workspaceService,
		TokensService:             tokensService,
		Listener:                  sqlListener,
		DynamicCredentialsService: dynamiccredsService,
		HostnameService:           hostnameService,
	})
	authenticatorService, err := authenticator.NewAuthenticatorService(ctx, authenticator.Options{
		Logger:               logger,
		URLClient:            hostnameService,
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

	sshkeyService := sshkey.NewService(sshkey.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Responder:  responder,
	})

	serverRunner, err := runner.New(
		logger,
		runnerService,
		func(_ string) runner.OperationClient {
			return runner.NewOperationClient(
				runService,
				workspaceService,
				variableService,
				stateService,
				configService,
				hostnameService,
				runnerService,
				sshkeyService,
			)
		},
		cfg.RunnerConfig,
	)
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Listener:   sqlListener,
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
		repoService,
		authenticatorService,
		loginserver.NewServer(loginserver.Options{
			Secret:      cfg.Secret,
			UserService: userService,
		}),
		configService,
		notificationService,
		runnerService,
		disco.Service{},
		&tfeapi.Handlers{},
		ui.NewHandlers(
			logger,
			runService,
			workspaceService,
			userService,
			teamService,
			orgService,
			moduleService,
			vcsService,
			stateService,
			runnerService,
			githubAppService,
			engineService,
			configService,
			hostnameService,
			tokensService,
			authorizer,
			authenticatorService,
			variableService,
			cfg.GithubHostname,
			sshkeyService,
			cfg.SkipTLSVerification,
			cfg.SiteToken,
			cfg.RestrictOrganizationCreation,
		),
		&github.AppEventHandler{
			Logger:     logger,
			Publisher:  vcsEventBroker,
			GithubApps: githubAppService,
			VCSService: vcsService,
		},
		dynamiccredsService,
		sshkeyService,
	}

	// Construct subsystems; ordered by start order
	subsystems := []*Subsystem{
		// The listener is started first because it is responsible for listening
		// for database events, and other subsystems rely on it to be listening
		// before they start, such as the runner (which generates a new runner
		// event), and the allocator (which receives the new runner event).
		{
			Name:   "listener",
			Logger: logger,
			System: sqlListener,
		},
		{
			Name:   "reporter",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.ReporterLockID),
			System: run.NewReporter(
				logger,
				vcsService,
				workspaceService,
				runService,
				configService,
				hostnameService,
			),
		},
		{
			Name:   "run_metrics",
			Logger: logger,
			System: runService.MetricsCollector,
		},
		{
			Name:   "timeout",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.TimeoutLockID),
			System: &run.Timeout{
				Logger:                logger.WithValues("component", "timeout"),
				OverrideCheckInterval: cfg.OverrideTimeoutCheckInterval,
				PlanningTimeout:       cfg.PlanningTimeout,
				ApplyingTimeout:       cfg.ApplyingTimeout,
				Runs:                  runService,
			},
		},
		{
			Name:   "run-deleter",
			Logger: logger,
			System: &resource.Deleter[*run.Run]{
				Logger:                logger.WithValues("component", "run-deleter"),
				OverrideCheckInterval: cfg.OverrideDeleterInterval,
				Client:                &runDeleterAdapter{svc: runService},
				AgeThreshold:          cfg.DeleteRunsAfter,
			},
		},
		{
			Name:   "config-deleter",
			Logger: logger,
			System: &resource.Deleter[*configversion.ConfigurationVersion]{
				Logger:                logger.WithValues("component", "config-deleter"),
				OverrideCheckInterval: cfg.OverrideDeleterInterval,
				Client:                &configVersionDeleterAdapter{svc: configService},
				AgeThreshold:          cfg.DeleteConfigsAfter,
			},
		},
		{
			Name:   "notifier",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.NotifierLockID),
			System: notifications.NewNotifier(notifications.NotifierOptions{
				Logger:             logger,
				HostnamesClient:    hostnameService,
				WorkspaceClient:    workspaceService,
				RunClient:          runService,
				NotificationClient: notificationService,
				DB:                 db,
			}),
		},
		{
			Name:   "job-allocator",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.AllocatorLockID),
			System: runnerService.NewAllocator(logger),
		},
		{
			Name:   "runner-manager",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.RunnerManagerLockID),
			System: runnerService.NewManager(),
		},
		{
			Name:   "job-signaler",
			Logger: logger,
			DB:     db,
			System: runnerService.Signaler,
		},
	}
	if !cfg.DisableRunner {
		subsystems = append(subsystems, &Subsystem{
			Name:   "runner-daemon",
			Logger: logger,
			DB:     db,
			System: serverRunner,
		})
	}
	if !cfg.DisableScheduler {
		subsystems = append(subsystems, &Subsystem{
			Name:   "scheduler",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.SchedulerLockID),
			System: run.NewScheduler(run.SchedulerOptions{
				Logger:          logger,
				WorkspaceClient: workspaceService,
				RunClient:       runService,
			}),
		})
	}

	// Construct web server and start listening on port
	server, err := http.NewServer(logger, http.ServerConfig{
		SSL:                  cfg.SSL,
		CertFile:             cfg.CertFile,
		KeyFile:              cfg.KeyFile,
		EnableRequestLogging: cfg.EnableRequestLogging,
		Middleware:           []mux.MiddlewareFunc{tokensService.Middleware()},
		Handlers:             handlers,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up http server: %w", err)
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
		SSHKeys:       sshkeyService,
		DB:            db,
		runner:        serverRunner,
		sqlListener:   sqlListener,
		netListener:   netListener,
		server:        server,
		subsystems:    subsystems,
	}, nil
}

// Start the otfd daemon and block until ctx is cancelled or an error is
// returned. The started channel is closed once the daemon has started.
func (d *Daemon) Start(ctx context.Context, started chan struct{}) error {
	// Cancel context the first time a func started with g.Go() fails
	g, ctx := errgroup.WithContext(ctx)

	for _, ss := range d.subsystems {
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
		if err := d.server.Start(ctx, d.netListener); err != nil {
			return fmt.Errorf("http server terminated: %w", err)
		}
		return nil
	})

	// Inform the caller the daemon has started
	close(started)

	// Block until error or Ctrl-C received.
	return g.Wait()
}

// Close releases daemon resources.
func (d *Daemon) Close() {
	// close active connections to http server
	d.netListener.Close()
	// close all db connections upon exit
	d.DB.Close()
}

// configVersionDeleterAdapter adapts configversion.Service to the
// resource.deleterClient interface.
type configVersionDeleterAdapter struct {
	svc *configversion.Service
}

func (a *configVersionDeleterAdapter) ListOlderThan(ctx context.Context, age time.Time) ([]*configversion.ConfigurationVersion, error) {
	return a.svc.ListConfigVersionsOlderThan(ctx, age)
}

func (a *configVersionDeleterAdapter) Delete(ctx context.Context, id resource.TfeID) error {
	return a.svc.DeleteConfigVersion(ctx, id)
}

// runDeleterAdapter adapts run.Service to the resource.deleterClient interface.
type runDeleterAdapter struct {
	svc *run.Service
}

func (a *runDeleterAdapter) ListOlderThan(ctx context.Context, age time.Time) ([]*run.Run, error) {
	return a.svc.ListRunsOlderThan(ctx, age)
}

func (a *runDeleterAdapter) Delete(ctx context.Context, id resource.TfeID) error {
	return a.svc.DeleteRun(ctx, id)
}
