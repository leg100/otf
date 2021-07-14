package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leg100/ots/app"
	cmdutil "github.com/leg100/ots/cmd"
	"github.com/leg100/ots/http"
	"github.com/leg100/ots/sqlite"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultAddress = ":8080"
	DefaultDBPath  = "ots.db"
)

var (
	DBPath string
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

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(os.Args[1:]); err != nil {
		panic(err.Error())
	}

	if server.SSL {
		if server.CertFile == "" || server.KeyFile == "" {
			fmt.Fprintf(os.Stderr, "must provide both --cert-file and --key-file")
			os.Exit(1)
		}
	}

	db, err := sqlite.New(DBPath, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
	server.StateVersionService = app.NewStateVersionService(stateVersionStore, server.WorkspaceService)
	server.ConfigurationVersionService = app.NewConfigurationVersionService(configurationVersionStore, server.WorkspaceService)
	server.RunService = app.NewRunService(runStore, server.WorkspaceService, server.ConfigurationVersionService)
	server.PlanService = app.NewPlanService(runStore)
	server.ApplyService = app.NewApplyService(runStore)

	// Run poller in background
	// agent := agent.NewAgent(
	// 	server.ConfigurationVersionService,
	// 	server.StateVersionService,
	// 	server.PlanService,
	// 	server.RunService,
	// )
	// go agent.Poller(ctx)

	if err := server.Open(); err != nil {
		server.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Server listening on %s\n", server.Addr)

	// Block until Ctrl-C received.
	if err := server.Wait(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
