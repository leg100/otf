package main

import (
	"fmt"
	"net"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type (
	services struct {
		organization.OrganizationService
		auth.AgentTokenService
		variable.VariableService
		state.StateService
		workspace.WorkspaceService
		otf.HostnameService
		configversion.ConfigurationVersionService
		auth.RegistrySessionService
		run.RunService
		repo.RepoService
		logs.LogsService
	}

	config struct {
		CloudConfigs []cloud.Config

		agentConfig          *agent.Config
		loggerConfig         *cmdutil.LoggerConfig
		cacheConfig          *inmem.CacheConfig
		secret               string // secret for signing URLs
		siteToken            string
		hostname             string
		address, database    string
		maxConfigSize        int64
		SSL                  bool
		CertFile, KeyFile    string
		EnableRequestLogging bool
		DevMode              bool
	}
)

func start(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Setup logger
	logger, err := cmdutil.NewLogger(d.LoggerConfig)
	if err != nil {
		return err
	}

	if d.DevMode {
		logger.Info("enabled developer mode")
	}

	// Group several daemons and if any one of them errors then terminate them
	// all
	g, ctx := errgroup.WithContext(ctx)

	// give local agent unlimited access to services
	agentCtx := otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "local-agent"})
	// give other components unlimited access too
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "app-user"})

	// Setup database connection
	db, err := sql.New(ctx, sql.Options{
		Logger:     logger,
		ConnString: d.database,
	})
	if err != nil {
		return err
	}
	defer db.Close()

	// Build list of http handlers for each service
	var handlers []otf.Handlers

	handlers = append(handlers, orgService)
	handlers = append(handlers, authService)

	// configure http server to use authentication middleware
	d.Middleware = append(d.Middleware, authService.TokenMiddleware)
	d.Middleware = append(d.Middleware, authService.SessionMiddleware)

	services, err := newServices(logger, db, cfg)
	if err != nil {
		return err
	}

	handlers = append(handlers, vcsProviderService)
	handlers = append(handlers, workspaceService)
	handlers = append(handlers, configService)
	handlers = append(handlers, runService)
	handlers = append(handlers, logsService)
	handlers = append(handlers, moduleService)
	handlers = append(handlers, stateService)
	handlers = append(handlers, variableService)

	// Setup and start http server
	server, err := http.NewServer(logger, *d.ServerConfig, handlers...)
	if err != nil {
		return fmt.Errorf("setting up http server: %w", err)
	}
	ln, err := net.Listen("tcp", d.address)
	if err != nil {
		return err
	}
	defer ln.Close()

	// If --hostname not set then use http server's listening address
	if cfg.Hostname == "" {
		services.SetHostname(otf.NormalizeAddress(ln.Addr().(*net.TCPAddr)))
	}
	a.V(0).Info("set system hostname", "hostname", a.hostname)

	services.StartExpirer(ctx)

	// Setup internal agent for processing runs
	agentClient := client.LocalClient{
		AgentTokenService:           authService,
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		VariableService:             variableService,
		StateService:                stateService,
		HostnameService:             hostnameService,
		ConfigurationVersionService: configService,
		RegistrySessionService:      authService,
		RunService:                  runService,
		LogsService:                 logsService,
	}
	agent, err := agent.NewAgent(
		logger.WithValues("component", "agent"),
		agentClient,
		*d.Config)
	if err != nil {
		return fmt.Errorf("initializing agent: %w", err)
	}

	// Run pubsub broker
	g.Go(func() error { return broker.Start(ctx) })

	// Run scheduler - if there is another scheduler running already then
	// this'll wait until the other scheduler exits.
	g.Go(func() error {
		return scheduler.Start(ctx, scheduler.Options{
			Logger:           logger,
			WorkspaceService: workspaceService,
			RunService:       runService,
			DB:               db,
			Subscriber:       broker,
		})
	})

	// Run run-spawner
	g.Go(func() error {
		err := run.StartSpawner(ctx, run.SpawnerOptions{
			Logger:                      logger,
			ConfigurationVersionService: configService,
			WorkspaceService:            workspaceService,
			VCSProviderService:          vcsProviderService,
			RunService:                  runService,
			Subscriber:                  broker,
		})
		if err != nil {
			return fmt.Errorf("spawner terminated: %w", err)
		}
		return nil
	})

	// Run module publisher
	g.Go(func() error {
		err := module.StartPublisher(ctx, module.PublisherOptions{
			Logger:             logger,
			VCSProviderService: vcsProviderService,
			Subscriber:         broker,
		})
		if err != nil {
			return fmt.Errorf("module publisher terminated: %w", err)
		}
		return nil
	})

	// Run PR reporter - if there is another reporter running already then
	// this'll wait until the other reporter exits.
	g.Go(func() error {
		err := run.StartReporter(ctx, run.ReporterOptions{
			Logger:                      logger,
			ConfigurationVersionService: configService,
			WorkspaceService:            workspaceService,
			VCSProviderService:          vcsProviderService,
			HostnameService:             hostnameService,
			Subscriber:                  broker,
			DB:                          db,
		})
		if err != nil {
			return fmt.Errorf("reporter terminated: %w", err)
		}
		return nil
	})

	// Run local agent in background
	g.Go(func() error {
		if err := agent.Start(agentCtx); err != nil {
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

	// Block until error or Ctrl-C received.
	return g.Wait()
}
