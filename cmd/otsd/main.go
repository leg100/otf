package main

import (
	"context"
	"fmt"
	"os"

	cmdutil "github.com/leg100/ots/cmd"
	"github.com/leg100/ots/http"
	"github.com/leg100/ots/sqlite"
	"github.com/spf13/cobra"
	driver "gorm.io/driver/sqlite"
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

	db, err := gorm.Open(driver.Open(DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err.Error())
	}

	server.OrganizationService = sqlite.NewOrganizationService(db)
	server.WorkspaceService = sqlite.NewWorkspaceService(db)
	server.StateVersionService = sqlite.NewStateVersionService(db)
	server.ConfigurationVersionService = sqlite.NewConfigurationVersionService(db)

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
