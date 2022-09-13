package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultAddress  = ":8080"
	DefaultDatabase = "postgres:///otf?host=/var/run/postgresql"
	DefaultDataDir  = "~/.otf-data"
)

// dbConnStr is the postgres connection string
var dbConnStr string

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := run(ctx, os.Args[1:]); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	// all calls to services are made as the privileged app user
	ctx = otf.AddSubjectToContext(ctx, &otf.AppUser{})

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

	cmd.Flags().StringVar(&dbConnStr, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")

	loggerCfg := cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cacheCfg := newCacheConfigFromFlags(cmd.Flags())
	serverCfg := newServerConfigFromFlags(cmd.Flags())

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
	logger, err := cmdutil.NewLogger(loggerCfg)
	if err != nil {
		return err
	}

	// Setup cache
	cache, err := inmem.NewCache(*cacheCfg)
	if err != nil {
		return err
	}
	logger.Info("started cache", "max_size", cacheCfg.Size, "ttl", cacheCfg.TTL.String())

	// Setup database(s)
	db, err := sql.New(logger, dbConnStr, cache, sql.DefaultSessionCleanupInterval)
	if err != nil {
		return err
	}
	defer db.Close()

	// Group several daemons and if any one of them errors then terminate them
	// all
	g, gctx := errgroup.WithContext(ctx)

	// Setup pub sub broker
	pubsub, err := sql.NewPubSub(logger, db.Pool())
	if err != nil {
		return fmt.Errorf("setting up pub sub broker")
	}
	g.Go(func() error {
		if err := pubsub.Start(gctx); err != nil {
			return fmt.Errorf("pubsub daemon terminated prematurely: %w", err)
		}
		return nil
	})

	// Setup application services
	app, err := app.NewApplication(ctx, logger, db, cache, pubsub)
	if err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

	// run scheduler
	scheduler, err := otf.NewScheduler(ctx, logger, app)
	if err != nil {
		return fmt.Errorf("starting scheduler: %w", err)
	}
	go scheduler.Start(ctx)

	// Run local agent in background
	agent, err := agent.NewAgent(
		ctx,
		logger.WithValues("component", "agent"),
		app,
		agent.NewAgentOptions{Mode: agent.InternalAgentMode})
	if err != nil {
		return fmt.Errorf("unable to start agent: %w", err)
	}
	g.Go(func() error { return agent.Start(gctx) })

	server, err := http.NewServer(logger, *serverCfg, app, db, cache)
	if err != nil {
		return fmt.Errorf("setting up http server: %w", err)
	}

	g.Go(func() error {
		if err := server.Open(gctx); err != nil {
			return fmt.Errorf("web server terminated prematurely: %w", err)
		}
		return nil
	})

	// Block until error or Ctrl-C received.
	return g.Wait()
}

// newLoggerConfigFromFlags adds flags to the given flagset, and, after the
// flagset is parsed by the caller, the flags populate the returned logger
// config.
func newCacheConfigFromFlags(flags *pflag.FlagSet) *inmem.CacheConfig {
	cfg := inmem.CacheConfig{}

	flags.IntVar(&cfg.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	flags.DurationVar(&cfg.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")

	return &cfg
}

// newLoggerConfigFromFlags adds flags to the given flagset, and, after the
// flagset is parsed by the caller, the flags populate the returned logger
// config.
func newServerConfigFromFlags(flags *pflag.FlagSet) *http.ServerConfig {
	cfg := http.ServerConfig{}

	flags.StringVar(&cfg.Addr, "address", DefaultAddress, "Listening address")
	flags.BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	flags.StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	flags.StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	flags.BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	flags.BoolVar(&cfg.ApplicationConfig.DevMode, "dev-mode", false, "Enable developer mode.")
	flags.StringVar(&cfg.ApplicationConfig.Github.ClientID, "github-client-id", "", "Github Client ID")
	flags.StringVar(&cfg.ApplicationConfig.Github.ClientSecret, "github-client-secret", "", "Github Client Secret")
	flags.StringVar(&cfg.ApplicationConfig.Github.Hostname, "github-hostname", "github.com", "Github hostname")
	flags.StringVar(&cfg.SiteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")

	return &cfg
}
