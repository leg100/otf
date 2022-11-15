package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
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
		// Define run func in order to enable cobra's default help functionality
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.SetOut(out)

	var help, version bool

	cmd.Flags().StringVar(&dbConnStr, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().BoolVarP(&version, "version", "v", false, "Print version of otfd")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")

	loggerCfg := cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cacheCfg := newCacheConfigFromFlags(cmd.Flags())
	serverCfg := newServerConfigFromFlags(cmd.Flags())
	cloudCfgs := newCloudConfigsFromFlags(cmd.Flags())
	htmlCfg := newHTMLConfigFromFlags(cmd.Flags(), cloudCfgs)
	agentCfg := agent.NewConfigFromFlags(cmd.Flags())

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	if err := cmd.ParseFlags(args); err != nil {
		return err
	}

	if help {
		if err := cmd.Help(); err != nil {
			return err
		}
		return nil
	}

	if version {
		fmt.Fprintln(cmd.OutOrStdout(), otf.Version)
		return nil
	}

	// Validate and update configs following flag parsing
	for _, cfg := range cloudCfgs {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid cloud config: %w", err)
		}
		if err := cfg.UpdateEndpoint(); err != nil {
			return fmt.Errorf("updating oauth endpoints: %w", err)
		}
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

	// Group several daemons and if any one of them errors then terminate them
	// all
	g, ctx := errgroup.WithContext(ctx)

	// create context for local agent, in order to identify all calls it makes
	agentCtx := otf.AddSubjectToContext(ctx, &otf.LocalAgent{})
	// all other calls to services are made as the privileged app user
	ctx = otf.AddSubjectToContext(ctx, &otf.AppUser{})

	// Setup database(s)
	db, err := sql.New(ctx, logger, dbConnStr, cache, sql.DefaultSessionCleanupInterval)
	if err != nil {
		return err
	}
	defer db.Close()

	// Setup pub sub broker
	pubsub, err := sql.NewPubSub(logger, db)
	if err != nil {
		return fmt.Errorf("setting up pub sub broker")
	}

	// Setup application services
	app, err := app.NewApplication(ctx, logger, db, cache, pubsub)
	if err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

	g.Go(func() error { return pubsub.Start(ctx) })

	// Run scheduler - if there is another scheduler running already then
	// this'll wait until the other scheduler exits.
	g.Go(func() error {
		return otf.ExclusiveScheduler(ctx, logger, app)
	})

	// Run local agent in background
	g.Go(func() error {
		agent, err := agent.NewAgent(
			logger.WithValues("component", "agent"),
			app,
			*agentCfg)
		if err != nil {
			return fmt.Errorf("initializing agent: %w", err)
		}
		if err := agent.Start(agentCtx); err != nil {
			return fmt.Errorf("agent terminated: %w", err)
		}
		return nil
	})

	// Run HTTP/JSON-API server and web app
	g.Go(func() error {
		server, err := http.NewServer(logger, *serverCfg, app, db, cache)
		if err != nil {
			return fmt.Errorf("setting up http server: %w", err)
		}
		// add Web App routes
		if err := html.AddRoutes(logger, htmlCfg, serverCfg, app, server.Router); err != nil {
			return err
		}

		if err := server.Open(ctx); err != nil {
			return fmt.Errorf("web server terminated: %w", err)
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

	flags.StringVar(&cfg.Addr, "address", DefaultAddress, "Listening address")
	flags.BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	flags.StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	flags.StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	flags.BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	flags.StringVar(&cfg.SiteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	flags.StringVar(&cfg.Secret, "secret", "", "Secret string for signing short-lived URLs. Required.")

	return &cfg
}

// newCloudConfigFromFlags binds flags to web app config
func newHTMLConfigFromFlags(flags *pflag.FlagSet, cloudConfigs []*otf.CloudConfig) *html.Config {
	cfg := html.Config{
		CloudConfigs: make(map[otf.CloudName]*otf.CloudConfig),
	}

	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	for _, cc := range cloudConfigs {
		cfg.CloudConfigs[cc.Name] = cc
	}

	return &cfg
}

// newCloudConfigsFromFlags binds flags to cloud configs
func newCloudConfigsFromFlags(flags *pflag.FlagSet) []*otf.CloudConfig {
	cloudConfigs := []*otf.CloudConfig{
		otf.GithubDefaultConfig(),
		otf.GitlabDefaultConfig(),
	}

	for _, cc := range cloudConfigs {
		nameStr := string(cc.Name)
		flags.StringVar(&cc.ClientID, nameStr+"-client-id", "", nameStr+" client ID")
		flags.StringVar(&cc.ClientSecret, nameStr+"-client-secret", "", nameStr+" client secret")

		flags.StringVar(&cc.Hostname, nameStr+"-hostname", cc.Hostname, nameStr+" hostname")
		flags.BoolVar(&cc.SkipTLSVerification, nameStr+"-skip-tls-verification", false, "Skip "+nameStr+" TLS verification")
	}

	return cloudConfigs
}
