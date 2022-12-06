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
	var dbConnStr, hostname string

	cmd.Flags().StringVar(&dbConnStr, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&hostname, "hostname", DefaultAddress, "Hostname via which otfd is accessed")
	cmd.Flags().BoolVarP(&version, "version", "v", false, "Print version of otfd")
	cmd.Flags().BoolVarP(&help, "help", "h", false, "Print usage information")

	loggerCfg := cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cacheCfg := newCacheConfigFromFlags(cmd.Flags())
	serverCfg := newServerConfigFromFlags(cmd.Flags())
	htmlCfg := newHTMLConfigFromFlags(cmd.Flags())
	agentCfg := agent.NewConfigFromFlags(cmd.Flags())
	cloudCfgs := cloudFlags(cmd.Flags())

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

	// create oauth clients
	var oauthClients []*html.OAuthClient
	for _, cfg := range cloudCfgs {
		if cfg.ClientID == "" && cfg.ClientSecret == "" {
			// skip creating oauth client where creds are unspecified
			continue
		}
		client, err := html.NewOAuthClient(html.OAuthClientConfig{
			OTFHost:     hostname,
			CloudConfig: cfg.CloudConfig,
			Config:      cfg.Config,
		})
		if err != nil {
			return err
		}
		oauthClients = append(oauthClients, client)
	}

	// populate cloud service with cloud configurations
	var cloudServiceConfigs []otf.CloudConfig
	for _, cc := range cloudCfgs {
		cloudServiceConfigs = append(cloudServiceConfigs, cc.CloudConfig)
	}
	cloudService, err := inmem.NewCloudService(cloudServiceConfigs...)
	if err != nil {
		return err
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

	// give local agent unlimited access to services
	agentCtx := otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "local-agent"})
	// give other components unlimited access too
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "app-user"})

	// Setup database(s)
	db, err := sql.New(ctx, sql.Options{
		Logger:          logger,
		ConnString:      dbConnStr,
		Cache:           cache,
		CleanupInterval: sql.DefaultSessionCleanupInterval,
		CloudService:    cloudService,
	})
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
	app, err := app.NewApplication(ctx, app.Options{
		Logger:       logger,
		DB:           db,
		Cache:        cache,
		PubSub:       pubsub,
		CloudService: cloudService,
	})
	if err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

	g.Go(func() error { return pubsub.Start(ctx) })

	// Run scheduler - if there is another scheduler running already then
	// this'll wait until the other scheduler exits.
	g.Go(func() error {
		return otf.ExclusiveScheduler(ctx, logger, app)
	})

	// Run PR reporter - if there is another reporter running already then
	// this'll wait until the other reporter exits.
	g.Go(func() error {
		return otf.ExclusiveReporter(ctx, logger, hostname, app)
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
		htmlCfg.ServerConfig = serverCfg
		htmlCfg.Application = app
		htmlCfg.Router = server.Router
		htmlCfg.OAuthClients = oauthClients
		if err := html.AddRoutes(logger, *htmlCfg); err != nil {
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
	flags.Int64Var(&cfg.MaxConfigSize, "max-config-size", otf.DefaultConfigMaxSize, "Maximum permitted configuration size in bytes.")

	return &cfg
}

// newCloudConfigFromFlags binds flags to web app config
func newHTMLConfigFromFlags(flags *pflag.FlagSet) *html.ApplicationOptions {
	cfg := html.ApplicationOptions{}
	flags.BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")
	return &cfg
}
