package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
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

	_ "github.com/leg100/otf/sql/migrations"
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

	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
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
		ServerConfig:       newServerConfigFromFlags(cmd.Flags()),
		CacheConfig:        newCacheConfigFromFlags(cmd.Flags()),
		LoggerConfig:       cmdutil.NewLoggerConfigFromFlags(cmd.Flags()),
		ApplicationOptions: newHTMLConfigFromFlags(cmd.Flags()),
		Config:             agent.NewConfigFromFlags(cmd.Flags()),
		OAuthConfigs:       cloudFlags(cmd.Flags()),
	}
	cmd.RunE = d.run

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

func (d *daemon) run(cmd *cobra.Command, _ []string) error {
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
	hub, err := pubsub.NewHub(logger, pubsub.HubConfig{
		PoolDB: db,
	})
	if err != nil {
		return fmt.Errorf("setting up pub sub broker")
	}

	// Setup url signer
	signer := otf.NewSigner(d.secret)

	// Authorizer for site-wide access
	siteAuthorizer := otf.NewSiteAuthorizer(logger)

	// Web app template renderer
	renderer, err := html.NewViewEngine(d.DevMode)
	if err != nil {
		return fmt.Errorf("setting up renderer: %w", err)
	}

	// Build list of http handlers for each service
	var handlers []otf.Handlers

	orgService := organization.NewService(organization.Options{
		SiteAuthorizer: siteAuthorizer,
		DB:             db,
		Logger:         logger,
		PubSubService:  hub,
		Renderer:       renderer,
	})
	handlers = append(handlers, orgService)

	authService, err := auth.NewService(ctx, auth.Options{
		OrganizationAuthorizer: orgService,
		Configs:                d.OAuthConfigs,
		SiteToken:              d.siteToken,
	})
	handlers = append(handlers, authService)

	workspaceService := workspace.NewService(workspace.Options{
		TokenMiddleware:        authService.TokenMiddleware,
		SessionMiddleware:      authService.SessionMiddleware,
		OrganizationAuthorizer: orgService,
		DB:                     db,
		Logger:                 logger,
		PubSubService:          hub,
		Renderer:               renderer,
	})
	handlers = append(handlers, workspaceService)

	configService := configversion.NewService(configversion.Options{
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
		Database:            db,
		Signer:              signer,
		Logger:              logger,
		MaxUploadSize:       d.maxConfigSize,
	})
	handlers = append(handlers, configService)

	stateService := state.NewService(state.ServiceOptions{
		WorkspaceAuthorizer: workspaceService,
		Logger:              logger,
		Database:            db,
		Cache:               cache,
	})
	handlers = append(handlers, stateService)

	variableService := variable.NewService(variable.Options{
		WorkspaceAuthorizer: workspaceService,
		Logger:              logger,
		Database:            db,
		WorkspaceService:    workspaceService,
		Renderer:            renderer,
	})
	handlers = append(handlers, variableService)

	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		OrganizationAuthorizer: orgService,
		Service:                cloudService,
		DB:                     db,
		Renderer:               renderer,
		Logger:                 logger,
	})
	handlers = append(handlers, variableService)

	// Setup application services
	app, err := app.NewApplication(ctx, app.Options{
		Logger:              logger,
		DB:                  db,
		Cache:               cache,
		PubSub:              hub,
		CloudService:        cloudService,
		Authorizer:          authorizer,
		StateVersionService: stateService,
	})
	if err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

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

	d.ApplicationOptions.ServerConfig = d.ServerConfig
	d.ApplicationOptions.CloudConfigs = d.OAuthConfigs
	d.ApplicationOptions.VariableService = variableService

	// Setup client app for use by agent
	client := struct {
		otf.OrganizationService
		otf.AgentTokenService
		otf.VariableApp
		otf.StateVersionApp
		otf.WorkspaceService
		otf.HostnameService
		otf.ConfigurationVersionService
		otf.RegistrySessionService
		otf.RunService
		otf.EventService
	}{
		AgentTokenService:           app,
		WorkspaceService:            app.WorkspaceService,
		OrganizationService:         app,
		VariableApp:                 variableService,
		StateVersionApp:             stateService,
		HostnameService:             app,
		ConfigurationVersionService: app,
		RegistrySessionService:      registrySessionService,
		RunService:                  app,
		EventService:                app,
	}

	// Setup agent
	agent, err := agent.NewAgent(
		logger.WithValues("component", "agent"),
		client,
		*d.Config)
	if err != nil {
		return fmt.Errorf("initializing agent: %w", err)
	}

	// Run pubsub broker
	g.Go(func() error { return hub.Start(ctx) })

	// Run scheduler - if there is another scheduler running already then
	// this'll wait until the other scheduler exits.
	g.Go(func() error {
		return scheduler.ExclusiveScheduler(ctx, logger, app)
	})

	// Run triggerer
	g.Go(func() error {
		if err := triggerer.Start(ctx); err != nil {
			return fmt.Errorf("triggerer terminated: %w", err)
		}
		return nil
	})

	// Run PR reporter - if there is another reporter running already then
	// this'll wait until the other reporter exits.
	g.Go(func() error {
		return otf.ExclusiveReporter(ctx, logger, d.hostname, app)
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

	return &cfg
}

// newCloudConfigFromFlags binds flags to web app config
func newHTMLConfigFromFlags(flags *pflag.FlagSet) *html.ApplicationOptions {
	cfg := html.ApplicationOptions{}
	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")
	return &cfg
}
