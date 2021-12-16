package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
	"github.com/spf13/cobra"
)

const (
	DefaultAddress  = ":8080"
	DefaultDatabase = "postgres:///otf?host=/var/run/postgresql"
	DefaultDataDir  = "~/.otf-data"
	DefaultLogLevel = "info"
)

var (
	// database is the postgres connection string
	database string
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.HiRedString("Error:"), err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	server := http.NewServer()

	cacheConfig := inmem.CacheConfig{}

	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		// Define run func in order to enable cobra's default help functionality
		Run: func(cmd *cobra.Command, args []string) {},
	}

	var help bool

	cmd.Flags().StringVar(&server.Addr, "address", DefaultAddress, "Listening address")
	cmd.Flags().BoolVar(&server.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&server.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&server.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().StringVar(&database, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().BoolVar(&server.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	cmd.Flags().IntVar(&cacheConfig.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cacheConfig.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")
	cmd.Flags().BoolVar(&server.ApplicationConfig.DevMode, "dev-mode", false, "Enable developer mode.")

	loggerCfg := newLoggerConfigFromFlags(cmd.Flags())

	cmd.Flags().StringVar(&server.ApplicationConfig.GithubClientID, "github-client-id", "", "Github Client ID")
	cmd.MarkFlagRequired("github-client-id")
	cmd.Flags().StringVar(&server.ApplicationConfig.GithubClientSecret, "github-client-secret", "", "Github Client Secret")
	cmd.MarkFlagRequired("github-client-secret")
	cmd.Flags().StringVar(&server.ApplicationConfig.GithubRedirectURL, "github-redirect-url", html.DefaultGithubRedirectURL, "Github redirect URL")
	cmd.MarkFlagRequired("github-redirect-url")

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		return err
	}

	if help {
		if err := cmd.Help(); err != nil {
			return err
		}
		return nil
	}

	// Setup logger
	logger, err := newLogger(loggerCfg)
	if err != nil {
		return err
	}
	server.Logger = logger

	// Validate SSL params
	if server.SSL {
		if server.CertFile == "" || server.KeyFile == "" {
			return fmt.Errorf("must provide both --cert-file and --key-file")
		}
	}

	// Setup cache
	cache, err := inmem.NewCache(cacheConfig)
	if err != nil {
		return err
	}
	logger.Info("started cache", "max_size", cacheConfig.Size, "ttl", cacheConfig.TTL.String())

	// Setup postgres connection
	db, err := sql.New(logger, database)
	if err != nil {
		return err
	}
	defer db.Close()

	organizationStore := sql.NewOrganizationDB(db)
	workspaceStore := sql.NewWorkspaceDB(db)
	stateVersionStore := sql.NewStateVersionDB(db)
	runStore := sql.NewRunDB(db)
	configurationVersionStore := sql.NewConfigurationVersionDB(db)

	eventService := inmem.NewEventService(logger)

	planLogStore, err := inmem.NewChunkProxy(cache, sql.NewPlanLogDB(db))
	if err != nil {
		return fmt.Errorf("unable to instantiate plan log store: %w", err)
	}

	applyLogStore, err := inmem.NewChunkProxy(cache, sql.NewApplyLogDB(db))
	if err != nil {
		return fmt.Errorf("unable to instantiate apply log store: %w", err)
	}

	server.OrganizationService = app.NewOrganizationService(organizationStore, logger, eventService)
	server.WorkspaceService = app.NewWorkspaceService(workspaceStore, logger, server.OrganizationService, eventService)
	server.StateVersionService = app.NewStateVersionService(stateVersionStore, logger, server.WorkspaceService, cache)
	server.ConfigurationVersionService = app.NewConfigurationVersionService(configurationVersionStore, logger, server.WorkspaceService, cache)
	server.RunService = app.NewRunService(runStore, logger, server.WorkspaceService, server.ConfigurationVersionService, eventService, planLogStore, applyLogStore, cache)
	server.PlanService = app.NewPlanService(runStore, planLogStore, logger, eventService, cache)
	server.ApplyService = app.NewApplyService(runStore, applyLogStore, logger, eventService, cache)
	server.EventService = eventService
	server.CacheService = cache

	if server.ApplicationConfig.DevMode {
		logger.Info("enabled developer mode")
	}

	scheduler, err := inmem.NewScheduler(server.WorkspaceService, server.RunService, eventService, logger)
	if err != nil {
		return fmt.Errorf("unable to start scheduler: %w", err)
	}

	// Run scheduler in background TODO: error not handled
	go scheduler.Start(ctx)

	// Run poller in background
	agent, err := agent.NewAgent(
		logger,
		server.ConfigurationVersionService,
		server.StateVersionService,
		server.RunService,
		server.PlanService,
		server.ApplyService,
		eventService,
	)
	if err != nil {
		return fmt.Errorf("unable to start agent: %w", err)
	}

	go agent.Start(ctx)

	// Setup HTTP routes TODO: call within http pkg
	server.SetupRoutes()

	if err := server.Open(); err != nil {
		// TODO: defer close
		server.Close()
		return err
	}

	logger.Info("started server", "address", server.Addr, "ssl", server.SSL)

	// Block until Ctrl-C received.
	if err := server.Wait(ctx); err != nil {
		return err
	}

	return nil
}
