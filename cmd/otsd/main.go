package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/leg100/ots/http"
	"github.com/leg100/ots/sqlite"
	driver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultAddress            = ":8080"
	EnvironmentVariablePrefix = "OTS_"
	DefaultDBPath             = "ots.db"
)

var (
	DBPath string
)

func main() {
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	server := http.NewServer()

	fs := flag.NewFlagSet("otsd", flag.ContinueOnError)
	fs.StringVar(&server.Addr, "address", DefaultAddress, "Listening address")
	fs.BoolVar(&server.SSL, "ssl", false, "Toggle SSL")
	fs.StringVar(&server.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	fs.StringVar(&server.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	fs.StringVar(&DBPath, "db-path", DefaultDBPath, "Path to SQLite database file")

	SetFlagsFromEnvVariables(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err.Error())
	}

	if server.SSL {
		if server.CertFile == "" || server.KeyFile == "" {
			fmt.Fprintf(os.Stderr, "must provide both -cert-file and -key-file")
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

// Each flag can also be set with an env variable whose name starts with `OTS_`.
func SetFlagsFromEnvVariables(fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		envVar := flagToEnvVarName(f)
		if val, present := os.LookupEnv(envVar); present {
			fs.Set(f.Name, val)
		}
	})
}

// Unset env vars prefixed with `OTS_`
func UnsetEtokVars() {
	for _, kv := range os.Environ() {
		parts := strings.Split(kv, "=")
		if strings.HasPrefix(parts[0], EnvironmentVariablePrefix) {
			if err := os.Unsetenv(parts[0]); err != nil {
				panic(err.Error())
			}
		}
	}
}

func flagToEnvVarName(f *flag.Flag) string {
	return fmt.Sprintf("%s%s", EnvironmentVariablePrefix, strings.Replace(strings.ToUpper(f.Name), "-", "_", -1))
}
