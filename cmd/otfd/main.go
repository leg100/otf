package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	defaultAddress  = ":8080"
	defaultDatabase = "postgres:///otf?host=/var/run/postgresql"
)

type (
	services struct {
		organization.OrganizationService
		auth.AuthService
		variable.VariableService
		vcsprovider.VCSProviderService
		state.StateService
		workspace.WorkspaceService
		module.ModuleService
		otf.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		repo.RepoService
		logs.LogsService
		pubsub.Broker
	}

	config struct {
		agentConfig          *agent.Config
		loggerConfig         *cmdutil.LoggerConfig
		cacheConfig          *inmem.CacheConfig
		github               cloud.CloudOAuthConfig
		gitlab               cloud.CloudOAuthConfig
		secret               string // secret for signing URLs
		siteToken            string
		hostname             string
		address, database    string
		maxConfigSize        int64
		SSL                  bool
		CertFile, KeyFile    string
		EnableRequestLogging bool
		DevMode              bool
	}
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
	// set config defaults
	cfg := config{
		github: cloud.CloudOAuthConfig{
			Config:      github.Defaults(),
			OAuthConfig: github.OAuthDefaults(),
		},
		gitlab: cloud.CloudOAuthConfig{
			Config:      gitlab.Defaults(),
			OAuthConfig: gitlab.OAuthDefaults(),
		},
	}

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
	cmd.Flags().StringVar(&cfg.address, "address", defaultAddress, "Listening address")
	cmd.Flags().StringVar(&cfg.database, "database", defaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&cfg.hostname, "hostname", "", "User-facing hostname for otf")
	cmd.Flags().StringVar(&cfg.siteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	cmd.Flags().StringVar(&cfg.secret, "secret", "", "Secret string for signing short-lived URLs. Required.")
	cmd.Flags().Int64Var(&cfg.maxConfigSize, "max-config-size", configversion.DefaultConfigMaxSize, "Maximum permitted configuration size in bytes.")

	cmd.Flags().IntVar(&cfg.cacheConfig.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cfg.cacheConfig.TTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")

	cmd.Flags().BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	cmd.Flags().BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	cmd.Flags().StringVar(&cfg.github.Hostname, "github-hostname", cfg.github.Hostname, "github hostname")
	cmd.Flags().BoolVar(&cfg.github.SkipTLSVerification, "github-skip-tls-verification", false, "Skip github TLS verification")
	cmd.Flags().StringVar(&cfg.github.OAuthConfig.ClientID, "github-client-id", "", "github client ID")
	cmd.Flags().StringVar(&cfg.github.OAuthConfig.ClientSecret, "github-client-secret", "", "github client secret")

	cmd.Flags().StringVar(&cfg.gitlab.Hostname, "gitlab-hostname", cfg.gitlab.Hostname, "gitlab hostname")
	cmd.Flags().BoolVar(&cfg.gitlab.SkipTLSVerification, "gitlab-skip-tls-verification", false, "Skip gitlab TLS verification")
	cmd.Flags().StringVar(&cfg.gitlab.OAuthConfig.ClientID, "gitlab-client-id", "", "gitlab client ID")
	cmd.Flags().StringVar(&cfg.gitlab.OAuthConfig.ClientSecret, "gitlab-client-secret", "", "gitlab client secret")

	cfg.loggerConfig = cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cfg.agentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

func runFunc(cfg *config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		// Give superuser privileges to all server processes
		ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "app-user"})
		// Cancel context the first time a func started with g.Go() fails
		g, ctx := errgroup.WithContext(ctx)

		logger, err := cmdutil.NewLogger(cfg.loggerConfig)
		if err != nil {
			return err
		}

		if cfg.DevMode {
			logger.Info("enabled developer mode")
		}

		db, err := sql.New(ctx, sql.Options{
			Logger:     logger,
			ConnString: cfg.database,
		})
		if err != nil {
			return err
		}
		defer db.Close()

		services, err := newServices(logger, db, *cfg)
		if err != nil {
			return err
		}

		// Setup and start http server
		server, err := http.NewServer(logger, http.ServerConfig{
			Middleware: []mux.MiddlewareFunc{
				auth.AuthenticateToken(services.AuthService),
				auth.AuthenticateSession(services.AuthService),
			},
			Handlers: []otf.Handlers{
				services.OrganizationService,
				services.AuthService,
				services.VCSProviderService,
				services.WorkspaceService,
				services.ConfigurationVersionService,
				services.RunService,
				services.LogsService,
				services.ModuleService,
				services.StateService,
				services.VariableService,
			},
		})
		if err != nil {
			return fmt.Errorf("setting up http server: %w", err)
		}
		ln, err := net.Listen("tcp", cfg.address)
		if err != nil {
			return err
		}
		defer ln.Close()

		// If --hostname not set then use http server's listening address as
		// hostname
		if cfg.hostname == "" {
			listenAddress := ln.Addr().(*net.TCPAddr)
			services.SetHostname(otf.NormalizeAddress(listenAddress))
		}
		logger.V(0).Info("set system hostname", "hostname", cfg.hostname)

		// Start purging sessions on a regular interval
		services.StartExpirer(ctx)

		agent, err := agent.NewAgent(
			logger.WithValues("component", "agent"),
			client.LocalClient{
				AgentTokenService:           services.AuthService,
				WorkspaceService:            services.WorkspaceService,
				OrganizationService:         services.OrganizationService,
				VariableService:             services.VariableService,
				StateService:                services.StateService,
				HostnameService:             services.HostnameService,
				ConfigurationVersionService: services.ConfigurationVersionService,
				RegistrySessionService:      services.AuthService,
				RunService:                  services.RunService,
				LogsService:                 services.LogsService,
			},
			*cfg.agentConfig,
		)
		if err != nil {
			return fmt.Errorf("initializing agent: %w", err)
		}

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

func newServices(logger logr.Logger, db otf.DB, cfg config) (*services, error) {
	hostnameService := otf.NewHostnameService(cfg.hostname)

	renderer, err := html.NewViewEngine(cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("setting up web page renderer: %w", err)
	}
	cloudService, err := inmem.NewCloudService(cfg.github.Config, cfg.gitlab.Config)
	if err != nil {
		return nil, err
	}
	cache, err := inmem.NewCache(*cfg.cacheConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("started cache", "max_size", cfg.cacheConfig.Size, "ttl", cfg.cacheConfig.TTL)

	broker, err := pubsub.NewBroker(logger, pubsub.BrokerConfig{
		PoolDB: db,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up pub sub broker: %w", err)
	}
	// Setup url signer
	signer := otf.NewSigner(cfg.secret)

	orgService := organization.NewService(organization.Options{
		Logger:   logger,
		DB:       db,
		Renderer: renderer,
		Broker:   broker,
	})
	authService, err := auth.NewService(auth.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Configs:             []cloud.CloudOAuthConfig{cfg.github, cfg.gitlab},
		SiteToken:           cfg.siteToken,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up auth service: %w", err)
	}
	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:       logger,
		DB:           db,
		Renderer:     renderer,
		CloudService: cloudService,
	})
	repoService := repo.NewService(repo.Options{
		Logger:             logger,
		DB:                 db,
		CloudService:       cloudService,
		HostnameService:    hostnameService,
		Publisher:          broker,
		VCSProviderService: vcsProviderService,
	})

	workspaceService := workspace.NewService(workspace.Options{
		Logger:      logger,
		DB:          db,
		Broker:      broker,
		Renderer:    renderer,
		RepoService: repoService,
		TeamService: authService,
	})
	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
		Signer:              signer,
		MaxUploadSize:       cfg.maxConfigSize,
	})
	runService := run.NewService(run.Options{
		Logger:                      logger,
		DB:                          db,
		Renderer:                    renderer,
		WorkspaceAuthorizer:         workspaceService,
		WorkspaceService:            workspaceService,
		ConfigurationVersionService: configService,
		Broker:                      broker,
		Cache:                       cache,
		Signer:                      signer,
	})
	logsService := logs.NewService(logs.Options{
		Logger:        logger,
		DB:            db,
		RunAuthorizer: runService,
		Cache:         cache,
		Broker:        broker,
		Verifier:      signer,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		DB:                 db,
		Renderer:           renderer,
		HostnameService:    hostnameService,
		VCSProviderService: vcsProviderService,
		Signer:             signer,
		RepoService:        repoService,
	})
	stateService := state.NewService(state.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
	})
	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
	})

	return &services{
		AuthService:                 authService,
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		VariableService:             variableService,
		StateService:                stateService,
		ModuleService:               moduleService,
		HostnameService:             hostnameService,
		ConfigurationVersionService: configService,
		RunService:                  runService,
		LogsService:                 logsService,
	}, nil
}
