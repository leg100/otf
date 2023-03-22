package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

	if err := parseFlags(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func parseFlags(ctx context.Context, args []string, out io.Writer) error {
	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       otf.Version,
		RunE:          start,
	}
	cmd.SetOut(out)

	var cfg config

	// TODO: rename --address to --listen
	cmd.Flags().StringVar(&cfg.address, "address", DefaultAddress, "Listening address")
	cmd.Flags().StringVar(&cfg.database, "database", DefaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&cfg.hostname, "hostname", "", "User-facing hostname for otf")
	cmd.Flags().StringVar(&cfg.siteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	cmd.Flags().StringVar(&cfg.secret, "secret", "", "Secret string for signing short-lived URLs. Required.")
	cmd.Flags().Int64Var(&cfg.maxConfigSize, "max-config-size", configversion.DefaultConfigMaxSize, "Maximum permitted configuration size in bytes.")

	cmd.Flags().IntVar(&cfg.cacheSize, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cfg.cacheTTL, "cache-expiry", otf.DefaultCacheTTL, "Cache entry TTL.")

	cmd.Flags().BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")
	cmd.Flags().BoolVar(&cfg.DevMode, "dev-mode", false, "Enable developer mode.")

	cfg.LoggerConfig = cmdutil.NewLoggerConfigFromFlags(cmd.Flags())
	cfg.agentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

func newServices(logger logr.Logger, db otf.DB, cfg config) (*services, error) {
	hostnameService := otf.NewHostnameService(cfg.hostname)

	renderer, err := html.NewViewEngine(cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("setting up web page renderer: %w", err)
	}
	cloudService, err := inmem.NewCloudService(cfg.CloudConfigs...)
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
		Configs:             d.OAuthConfigs,
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

}
