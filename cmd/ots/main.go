package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/leg100/ots/boltdb"
	"github.com/leg100/ots/http"
	bolt "go.etcd.io/bbolt"
)

const (
	DefaultAddress            = ":8000"
	EnvironmentVariablePrefix = "OTS_"
)

func main() {
	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	db, err := bolt.Open("ots.db", 0600, nil)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	server := http.NewServer()

	server.OrganizationService = boltdb.NewOrganizationService(db)

	fs := flag.NewFlagSet("ots", flag.ContinueOnError)
	fs.StringVar(&server.Addr, "address", DefaultAddress, "Listening address")

	SetFlagsFromEnvVariables(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err.Error())
	}

	if err := server.Open(); err != nil {
		server.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

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
