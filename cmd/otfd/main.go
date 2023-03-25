package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/services"
	"github.com/leg100/otf/sql"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	defaultAddress  = ":8080"
	defaultDatabase = "postgres:///otf?host=/var/run/postgresql"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := parseFlags(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func parseFlags(ctx context.Context, args []string, out io.Writer) error {
	cfg := services.NewDefaultConfig()

	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       otf.Version,
		RunE:          runFunc(&cfg),
	}
	cmd.SetOut(out)

	// TODO: rename --address to --listen
	cmd.Flags().StringVar(&cfg.Address, "address", defaultAddress, "Listening address")
	cmd.Flags().StringVar(&cfg.Database, "database", defaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&cfg.Hostname, "hostname", "", "User-facing hostname for otf")
	cmd.Flags().StringVar(&cfg.SiteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	cmd.Flags().StringVar(&cfg.Secret, "secret", "", "Secret string for signing short-lived URLs. Required.")
	cmd.Flags().Int64Var(&cfg.MaxConfigSize, "max-config-size", configversion.DefaultConfigMaxSize, "Maximum permitted configuration size in bytes.")

	cmd.Flags().IntVar(&cfg.CacheConfig.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cfg.CacheConfig.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")

	cmd.Flags().BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	cmd.Flags().BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	cmd.Flags().StringVar(&cfg.Github.Hostname, "github-hostname", cfg.Github.Hostname, "github hostname")
	cmd.Flags().BoolVar(&cfg.Github.SkipTLSVerification, "github-skip-tls-verification", false, "Skip github TLS verification")
	cmd.Flags().StringVar(&cfg.Github.OAuthConfig.ClientID, "github-client-id", "", "github client ID")
	cmd.Flags().StringVar(&cfg.Github.OAuthConfig.ClientSecret, "github-client-secret", "", "github client secret")

	cmd.Flags().StringVar(&cfg.Gitlab.Hostname, "gitlab-hostname", cfg.Gitlab.Hostname, "gitlab hostname")
	cmd.Flags().BoolVar(&cfg.Gitlab.SkipTLSVerification, "gitlab-skip-tls-verification", false, "Skip gitlab TLS verification")
	cmd.Flags().StringVar(&cfg.Gitlab.OAuthConfig.ClientID, "gitlab-client-id", "", "gitlab client ID")
	cmd.Flags().StringVar(&cfg.Gitlab.OAuthConfig.ClientSecret, "gitlab-client-secret", "", "gitlab client secret")

	cfg.LoggerConfig = cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cfg.AgentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

func runFunc(cfg *services.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		// Give superuser privileges to all server processes
		ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "app-user"})
		// Cancel context the first time a func started with g.Go() fails
		g, ctx := errgroup.WithContext(ctx)

		logger, err := cmdutil.NewLogger(cfg.LoggerConfig)
		if err != nil {
			return err
		}

		if cfg.DevMode {
			logger.Info("enabled developer mode")
		}

		db, err := sql.New(ctx, sql.Options{
			Logger:     logger,
			ConnString: cfg.Database,
		})
		if err != nil {
			return err
		}
		defer db.Close()

		services, err := services.New(logger, db, *cfg)
		if err != nil {
			return err
		}

		agent, err := services.NewAgent(logger)
		if err != nil {
			return err
		}

		server, err := http.NewServer(logger, http.ServerConfig{
			SSL:                  cfg.SSL,
			CertFile:             cfg.CertFile,
			KeyFile:              cfg.KeyFile,
			EnableRequestLogging: cfg.EnableRequestLogging,
			DevMode:              cfg.DevMode,
			Middleware: []mux.MiddlewareFunc{
				auth.AuthenticateToken(services.AuthService),
				auth.AuthenticateSession(services.AuthService),
			},
			Handlers: services.Handlers,
		})
		if err != nil {
			return fmt.Errorf("setting up http server: %w", err)
		}
		ln, err := net.Listen("tcp", cfg.Address)
		if err != nil {
			return err
		}
		defer ln.Close()

		// If --hostname not set then use http server's listening address as
		// hostname
		if cfg.Hostname == "" {
			listenAddress := ln.Addr().(*net.TCPAddr)
			services.SetHostname(otf.NormalizeAddress(listenAddress))
		}
		logger.V(0).Info("set system hostname", "hostname", cfg.Hostname)

		// Start purging sessions on a regular interval
		services.StartExpirer(ctx)

		// Run pubsub broker
		g.Go(func() error { return services.Broker.Start(ctx) })

		// Run scheduler - if there is another scheduler running already then
		// this'll wait until the other scheduler exits.
		g.Go(func() error {
			return scheduler.Start(ctx, scheduler.Options{
				Logger:           logger,
				WorkspaceService: services.WorkspaceService,
				RunService:       services.RunService,
				DB:               db,
				Subscriber:       services,
			})
		})

		// Run run-spawner
		g.Go(func() error {
			err := run.StartSpawner(ctx, run.SpawnerOptions{
				Logger:                      logger,
				ConfigurationVersionService: services.ConfigurationVersionService,
				WorkspaceService:            services.WorkspaceService,
				VCSProviderService:          services.VCSProviderService,
				RunService:                  services.RunService,
				Subscriber:                  services.Broker,
			})
			if err != nil {
				return fmt.Errorf("spawner terminated: %w", err)
			}
			return nil
		})

		// Run module publisher
		g.Go(func() error {
			err := module.StartPublisher(ctx, module.PublisherOptions{
				Logger:             logger,
				VCSProviderService: services.VCSProviderService,
				Subscriber:         services.Broker,
			})
			if err != nil {
				return fmt.Errorf("module publisher terminated: %w", err)
			}
			return nil
		})

		// Run PR reporter - if there is another reporter running already then
		// this'll wait until the other reporter exits.
		g.Go(func() error {
			err := run.StartReporter(ctx, run.ReporterOptions{
				Logger:                      logger,
				ConfigurationVersionService: services.ConfigurationVersionService,
				WorkspaceService:            services.WorkspaceService,
				VCSProviderService:          services.VCSProviderService,
				HostnameService:             services.HostnameService,
				Subscriber:                  services.Broker,
				DB:                          db,
			})
			if err != nil {
				return fmt.Errorf("reporter terminated: %w", err)
			}
			return nil
		})

		// Run local agent in background
		g.Go(func() error {
			// give local agent unlimited access to services
			ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "local-agent"})

			if err := agent.Start(ctx); err != nil {
				return fmt.Errorf("agent terminated: %w", err)
			}
			return nil
		})

		// Run HTTP/JSON-API server and web app
		g.Go(func() error {
			if err := server.Start(ctx, ln); err != nil {
				return fmt.Errorf("http server terminated: %w", err)
			}
			return nil
		})

		// Block until error or Ctrl-C received.
		return g.Wait()
	}
}
