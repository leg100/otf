package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultAddress  = ":8080"
	DefaultDatabase = "postgres:///otf?host=/var/run/postgresql"
	DefaultDataDir  = "~/.otf-data"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := runDaemon(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func runDaemon(ctx context.Context, args []string, out io.Writer) error {
	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       otf.Version,
	}
	cmd.SetOut(out)

	d := &daemon{
		ServerConfig: newServerConfigFromFlags(cmd.Flags()),
		CacheConfig:  newCacheConfigFromFlags(cmd.Flags()),
		LoggerConfig: cmdutil.NewLoggerConfigFromFlags(cmd.Flags()),
		Config:       agent.NewConfigFromFlags(cmd.Flags()),
		OAuthConfigs: cloudFlags(cmd.Flags()),
	}
	cmd.RunE = d.start

	// TODO: rename --address to --listen
	cmd.Flags().StringVar(&d.address, "address", DefaultAddress, "Listening address")
	cmd.Flags().StringVar(&d.database, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&d.hostname, "hostname", "", "User-facing hostname for otf")
	cmd.Flags().StringVar(&d.siteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	cmd.Flags().StringVar(&d.secret, "secret", "", "Secret string for signing short-lived URLs. Required.")
	cmd.Flags().Int64Var(&d.maxConfigSize, "max-config-size", configversion.DefaultConfigMaxSize, "Maximum permitted configuration size in bytes.")

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

type daemon struct {
	address, hostname, database, siteToken, secret string
	maxConfigSize                                  int64

	*http.ServerConfig
	*inmem.CacheConfig
	*cmdutil.LoggerConfig
	*agent.Config
	cloud.OAuthConfigs
}

func (d *daemon) start(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Setup logger
	logger, err := cmdutil.NewLogger(d.LoggerConfig)
	if err != nil {
		return err
	}

	if d.DevMode {
		logger.Info("enabled developer mode")
	}

	// Setup hostname service
	hostnameService := otf.NewHostnameService(logger)

	// populate cloud service with cloud configurations
	cloudService, err := inmem.NewCloudService(d.OAuthConfigs.Configs()...)
	if err != nil {
		return err
	}

	// Setup cache
	cache, err := inmem.NewCache(*d.CacheConfig)
	if err != nil {
		return err
	}
	logger.Info("started cache", "max_size", d.CacheConfig.Size, "ttl", d.CacheConfig.TTL.String())

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

	// Setup pub sub broker
	broker, err := pubsub.NewBroker(logger, pubsub.BrokerConfig{
		PoolDB: db,
	})
	if err != nil {
		return fmt.Errorf("setting up pub sub broker")
	}

	// Setup url signer
	signer := otf.NewSigner(d.secret)

	// Web app template renderer
	renderer, err := html.NewViewEngine(d.DevMode)
	if err != nil {
		return fmt.Errorf("setting up renderer: %w", err)
	}

	// Build list of http handlers for each service
	var handlers []otf.Handlers

	orgService := organization.NewService(organization.Options{
		Logger:    logger,
		DB:        db,
		Renderer:  renderer,
		Publisher: broker,
	})
	handlers = append(handlers, orgService)

	authService, err := auth.NewService(ctx, auth.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Configs:             d.OAuthConfigs,
		SiteToken:           d.siteToken,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
	})
	if err != nil {
		return fmt.Errorf("setting up auth service: %w", err)
	}
	handlers = append(handlers, authService)

	// configure http server to use authentication middleware
	d.Middleware = append(d.Middleware, authService.TokenMiddleware)
	d.Middleware = append(d.Middleware, authService.SessionMiddleware)

	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:       logger,
		DB:           db,
		Renderer:     renderer,
		CloudService: cloudService,
	})
	handlers = append(handlers, vcsProviderService)

	repoService := repo.NewService(repo.Options{
		Logger:             logger,
		DB:                 db,
		CloudService:       cloudService,
		HostnameService:    hostnameService,
		Publisher:          broker,
		VCSProviderService: vcsProviderService,
	})

	workspaceService := workspace.NewService(workspace.Options{
		Logger:      logger,
		DB:          db,
		Broker:      broker,
		Renderer:    renderer,
		RepoService: repoService,
		TeamService: authService,
	})
	handlers = append(handlers, workspaceService)

	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
		Signer:              signer,
		MaxUploadSize:       d.maxConfigSize,
	})
	handlers = append(handlers, configService)

	runService := run.NewService(run.Options{
		Logger:                      logger,
		DB:                          db,
		Renderer:                    renderer,
		WorkspaceAuthorizer:         workspaceService,
		WorkspaceService:            workspaceService,
		ConfigurationVersionService: configService,
		Broker:                      broker,
		Cache:                       cache,
		Signer:                      signer,
	})
	handlers = append(handlers, runService)

	logsService := logs.NewService(logs.Options{
		Logger:        logger,
		DB:            db,
		RunAuthorizer: runService,
		Cache:         cache,
		Broker:        broker,
		Verifier:      signer,
	})
	handlers = append(handlers, logsService)

	moduleService := module.NewService(module.Options{
		Logger:             logger,
		DB:                 db,
		Renderer:           renderer,
		CloudService:       cloudService,
		HostnameService:    hostnameService,
		VCSProviderService: vcsProviderService,
		Signer:             signer,
		RepoService:        repoService,
	})
	handlers = append(handlers, moduleService)

	stateService := state.NewService(state.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
	})
	handlers = append(handlers, stateService)

	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
	})
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

	// Set system hostname
	if err := hostnameService.SetHostname(d.hostname, ln.Addr().(*net.TCPAddr)); err != nil {
		return err
	}

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

// newCacheConfigFromFlags adds flags pertaining to cache config
func newCacheConfigFromFlags(flags *pflag.FlagSet) *inmem.CacheConfig {
	cfg := inmem.CacheConfig{}

	flags.IntVar(&cfg.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	flags.DurationVar(&cfg.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")

	return &cfg
}

// newServerConfigFromFlags adds flags pertaining to http server config
func newServerConfigFromFlags(flags *pflag.FlagSet) *http.ServerConfig {
	cfg := http.ServerConfig{}

	flags.BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	flags.StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	flags.StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	flags.BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	return &cfg
}
