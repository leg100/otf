package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
	"github.com/leg100/zerologr"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
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

	// Toggle log colors. Must be one of auto, true, or false.
	var logColor string

	cmd.Flags().StringVar(&server.Addr, "address", DefaultAddress, "Listening address")
	cmd.Flags().BoolVar(&server.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&server.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&server.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().StringVar(&database, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().BoolVar(&server.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	cmd.Flags().StringVar(&logColor, "log-color", "auto", "Toggle log colors: auto, true or false. Auto enables colors if using a TTY.")
	cmd.Flags().IntVar(&cacheConfig.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cacheConfig.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")
	logLevel := cmd.Flags().StringP("log-level", "l", DefaultLogLevel, "Logging level")

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		panic(err.Error())
	}

	if help {
		if err := cmd.Help(); err != nil {
			panic(err.Error())
		}
		os.Exit(0)
	}

	// Setup logger
	zerologger, err := newLogger(*logLevel, logColor)
	if err != nil {
		panic(err.Error())
	}
	logger := zerologr.NewLogger(zerologger)
	server.Logger = logger

	// Validate SSL params
	if server.SSL {
		if server.CertFile == "" || server.KeyFile == "" {
			fmt.Fprintf(os.Stderr, "must provide both --cert-file and --key-file")
			os.Exit(1)
		}
	}

	// Setup cache
	cache, err := inmem.NewCache(cacheConfig)
	if err != nil {
		panic(err.Error())
	}
	logger.Info("started cache", "max_size", cacheConfig.Size, "ttl", cacheConfig.TTL.String())

	// Setup postgres connection
	db, err := sql.New(logger, database, sql.WithZeroLogger(zerologger))
	if err != nil {
		panic(err.Error())
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
		panic(fmt.Sprintf("unable to instantiate plan log store: %s", err.Error()))
	}

	applyLogStore, err := inmem.NewChunkProxy(cache, sql.NewApplyLogDB(db))
	if err != nil {
		panic(fmt.Sprintf("unable to instantiate apply log store: %s", err.Error()))
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

	scheduler, err := inmem.NewScheduler(server.WorkspaceService, server.RunService, eventService, logger)
	if err != nil {
		panic(fmt.Sprintf("unable to start scheduler: %s", err.Error()))
	}

	// Run scheduler in background
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
		panic(fmt.Sprintf("unable to start agent: %s", err.Error()))
	}
	go agent.Start(ctx)

	// Setup HTTP routes
	server.SetupRoutes()

	if err := server.Open(); err != nil {
		server.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger.Info("started server", "address", server.Addr, "ssl", server.SSL)

	// Block until Ctrl-C received.
	if err := server.Wait(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newLogger(lvl, color string) (*zerolog.Logger, error) {
	zlvl, err := zerolog.ParseLevel(lvl)
	if err != nil {
		return nil, err
	}

	// Setup logger
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.DurationFieldInteger = true

	switch color {
	case "auto":
		// Disable color if stdout is not a tty
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			consoleWriter.NoColor = true
		}
	case "true":
		consoleWriter.NoColor = false
	case "false":
		consoleWriter.NoColor = true
	default:
		return nil, fmt.Errorf("invalid choice for log color: %s. Must be one of auto, true, or false", color)
	}

	logger := zerolog.New(consoleWriter).Level(zlvl).With().Timestamp().Logger()

	if logger.GetLevel() < zerolog.InfoLevel {
		// Inform the user that logging lower than INFO threshold has been
		// enabled
		logger.WithLevel(logger.GetLevel()).Msg("custom log level enabled")
	}

	return &logger, nil
}
