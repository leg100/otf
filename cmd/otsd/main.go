package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/butonic/zerologr"
	"github.com/leg100/ots/agent"
	"github.com/leg100/ots/app"
	cmdutil "github.com/leg100/ots/cmd"
	"github.com/leg100/ots/filestore"
	"github.com/leg100/ots/http"
	"github.com/leg100/ots/sqlite"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	DefaultAddress  = ":8080"
	DefaultHostname = "localhost:8080"
	DefaultDBPath   = "ots.db"
	DefaultDataDir  = "~/.ots-data"
)

var (
	// DBPath is the path to the sqlite database file
	DBPath string
	// DataDir is the path to the directory used for storing OTS-related data
	DataDir string
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	server := http.NewServer()

	cmd := &cobra.Command{
		Use:           "otsd",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVar(&server.Addr, "address", DefaultAddress, "Listening address")
	cmd.Flags().BoolVar(&server.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&server.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&server.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().StringVar(&DBPath, "db-path", DefaultDBPath, "Path to SQLite database file")
	cmd.Flags().StringVar(&server.Hostname, "hostname", DefaultHostname, "Hostname used within absolute URL links")
	cmd.Flags().StringVar(&DataDir, "data-dir", DefaultDataDir, "Path to directory for storing OTS related data")

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		panic(err.Error())
	}

	// DataDir: Expand ~ to home dir
	var err error
	DataDir, err = homedir.Expand(DataDir)
	if err != nil {
		panic(err.Error())
	}

	// Setup logger
	zerologger := newLogger()
	logger := zerologr.NewWithOptions(zerologr.Options{Logger: zerologger})
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
	logger.Info("filestore started", "path", fs.Path())

	// Setup sqlite db
	db, err := sqlite.New(DBPath, sqlite.WithZeroLogger(zerologger))
	if err != nil {
		panic(err.Error())
	}

	organizationStore := sqlite.NewOrganizationDB(db)
	workspaceStore := sqlite.NewWorkspaceDB(db)
	stateVersionStore := sqlite.NewStateVersionDB(db)
	runStore := sqlite.NewRunDB(db)
	configurationVersionStore := sqlite.NewConfigurationVersionDB(db)

	server.OrganizationService = app.NewOrganizationService(organizationStore)
	server.WorkspaceService = app.NewWorkspaceService(workspaceStore, server.OrganizationService)
	server.StateVersionService = app.NewStateVersionService(stateVersionStore, server.WorkspaceService, fs)
	server.ConfigurationVersionService = app.NewConfigurationVersionService(configurationVersionStore, server.WorkspaceService, fs)
	server.RunService = app.NewRunService(runStore, server.WorkspaceService, server.ConfigurationVersionService, fs)
	server.PlanService = app.NewPlanService(runStore, fs)
	server.ApplyService = app.NewApplyService(runStore)

	// Run poller in background
	agent := agent.NewAgent(
		logger,
		server.ConfigurationVersionService,
		server.StateVersionService,
		server.PlanService,
		server.RunService,
	)
	go agent.Poller(ctx)

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

func newLogger() *zerolog.Logger {
	// Setup logger
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.DurationFieldInteger = true

	logger := zerolog.New(consoleWriter).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	return &logger
}
