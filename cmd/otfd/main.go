package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/filestore"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
	"github.com/leg100/zerologr"
	"github.com/mitchellh/go-homedir"
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
	// DataDir is the path to the directory used for storing OTF-related data
	DataDir string
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	server := http.NewServer()

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
	cmd.Flags().StringVar(&DataDir, "data-dir", DefaultDataDir, "Path to directory for storing OTF related data")
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

	// DataDir: Expand ~ to home dir
	var err error
	DataDir, err = homedir.Expand(DataDir)
	if err != nil {
		panic(err.Error())
	}

	// Setup logger
	zerologger, err := newLogger(*logLevel)
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

	// Setup filestore
	fs, err := filestore.NewFilestore(DataDir)
	if err != nil {
		panic(err.Error())
	}
	logger.Info("filestore started", "path", fs.Path)

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

	server.OrganizationService = app.NewOrganizationService(organizationStore, logger, eventService)
	server.WorkspaceService = app.NewWorkspaceService(workspaceStore, logger, server.OrganizationService, eventService)
	server.StateVersionService = app.NewStateVersionService(stateVersionStore, logger, server.WorkspaceService, fs)
	server.ConfigurationVersionService = app.NewConfigurationVersionService(configurationVersionStore, logger, server.WorkspaceService, fs)
	server.RunService = app.NewRunService(runStore, logger, server.WorkspaceService, server.ConfigurationVersionService, fs, eventService)
	server.PlanService = app.NewPlanService(runStore, fs)
	server.ApplyService = app.NewApplyService(runStore)
	server.EventService = eventService

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

	logger.Info("server started", "address", server.Addr, "ssl", server.SSL)

	// Block until Ctrl-C received.
	if err := server.Wait(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newLogger(lvl string) (*zerolog.Logger, error) {
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

	logger := zerolog.New(consoleWriter).Level(zlvl).With().Timestamp().Logger()

	if logger.GetLevel() < zerolog.InfoLevel {
		// Inform the user that logging lower than INFO threshold has been
		// enabled
		logger.WithLevel(logger.GetLevel()).Msg("custom log level enabled")
	}

	return &logger, nil
}
