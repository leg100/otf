package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/app"
	"github.com/leg100/otf/cmd"
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
		Version:       otf.Version,
	}
	cmd.SetOut(out)

	d := &daemon{
		ServerConfig:       newServerConfigFromFlags(cmd.Flags()),
		CacheConfig:        newCacheConfigFromFlags(cmd.Flags()),
		LoggerConfig:       cmdutil.NewLoggerConfigFromFlags(cmd.Flags()),
		ApplicationOptions: newHTMLConfigFromFlags(cmd.Flags()),
		Config:             agent.NewConfigFromFlags(cmd.Flags()),
		cloudConfigs:       cloudFlags(cmd.Flags()),
	}
	cmd.RunE = d.run

	cmd.Flags().StringVar(&d.database, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&d.hostname, "hostname", DefaultAddress, "User-facing hostname for otf")

	cmdutil.SetFlagsFromEnvVariables(cmd.Flags())

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

type daemon struct {
	hostname string
	database string

	*http.ServerConfig
	*inmem.CacheConfig
	*cmd.LoggerConfig
	*html.ApplicationOptions
	*agent.Config
	cloudConfigs []*cloudConfig
}

func (d *daemon) run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	d.ServerConfig.Hostname = d.hostname

	// create oauth clients
	var oauthClients []*html.OAuthClient
	for _, cfg := range d.cloudConfigs {
		if cfg.ClientID == "" && cfg.ClientSecret == "" {
			// skip creating oauth client where creds are unspecified
			continue
		}
		client, err := html.NewOAuthClient(html.OAuthClientConfig{
			OTFHost:     d.hostname,
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
	for _, cc := range d.cloudConfigs {
		cloudServiceConfigs = append(cloudServiceConfigs, cc.CloudConfig)
	}
	cloudService, err := inmem.NewCloudService(cloudServiceConfigs...)
	if err != nil {
		return err
	}

	// Setup logger
	logger, err := cmdutil.NewLogger(d.LoggerConfig)
	if err != nil {
		return err
	}

	// Setup cache
	cache, err := inmem.NewCache(*d.CacheConfig)
	if err != nil {
		return err
	}
	logger.Info("started cache", "max_size", d.CacheConfig.Size, "ttl", d.CacheConfig.TTL.String())

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
		ConnString:      d.database,
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
		return otf.ExclusiveReporter(ctx, logger, d.hostname, app)
	})

	// Run local agent in background
	g.Go(func() error {
		d.Config.RegistryHostname = d.hostname

		agent, err := agent.NewAgent(
			logger.WithValues("component", "agent"),
			app,
			*d.Config)
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
		server, err := http.NewServer(logger, *d.ServerConfig, app, db, cache)
		if err != nil {
			return fmt.Errorf("setting up http server: %w", err)
		}
		d.ApplicationOptions.ServerConfig = d.ServerConfig
		d.ApplicationOptions.Application = app
		d.ApplicationOptions.Router = server.Router
		d.ApplicationOptions.OAuthClients = oauthClients
		if err := html.AddRoutes(logger, *d.ApplicationOptions); err != nil {
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

	// TODO: rename --address to --listen
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
