package main

import (
	"context"
	"io"
	"os"

	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/daemon"
	"github.com/leg100/otf/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	cfg := daemon.Config{}
	daemon.ApplyDefaults(&cfg)

	var loggerConfig *logr.Config

	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       otf.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logr.New(loggerConfig)
			if err != nil {
				return err
			}

			// Confer superuser privileges on all calls to service endpoints
			ctx := otf.AddSubjectToContext(cmd.Context(), &otf.Superuser{Username: "app-user"})

			d, err := daemon.New(ctx, logger, cfg)
			if err != nil {
				return err
			}
			// block until ^C received
			return d.Start(ctx, make(chan struct{}))
		},
	}
	cmd.SetOut(out)

	// TODO: rename --address to --listen
	cmd.Flags().StringVar(&cfg.Address, "address", defaultAddress, "Listening address")
	cmd.Flags().StringVar(&cfg.Database, "database", defaultDatabase, "Postgres connection string")
	cmd.Flags().StringVar(&cfg.Host, "hostname", "", "User-facing hostname for otf")
	cmd.Flags().StringVar(&cfg.SiteToken, "site-token", "", "API token with site-wide unlimited permissions. Use with care.")
	cmd.Flags().StringSliceVar(&cfg.SiteAdmins, "site-admins", nil, "Promote a list of users to site admin.")
	cmd.Flags().StringVar(&cfg.Secret, "secret", "", "Secret string for signing short-lived URLs. Required.")
	cmd.Flags().Int64Var(&cfg.MaxConfigSize, "max-config-size", cfg.MaxConfigSize, "Maximum permitted configuration size in bytes.")

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

	cmd.Flags().StringVar(&cfg.OIDC.Name, "oidc-name", cfg.OIDC.Name, "user friendly oidc name")
	cmd.Flags().StringVar(&cfg.OIDC.IssuerURL, "oidc-issuer-url", cfg.OIDC.IssuerURL, "oidc issuer url")
	cmd.Flags().StringVar(&cfg.OIDC.ClientID, "oidc-client-id", "", "oidc client ID")
	cmd.Flags().StringVar(&cfg.OIDC.ClientSecret, "oidc-client-secret", "", "oidc client secret")

	cmd.Flags().BoolVar(&cfg.RestrictOrganizationCreation, "restrict-org-creation", false, "Restrict organization creation capability to site admin role")

	cmd.Flags().StringVar(&cfg.GoogleIAPConfig.Audience, "google-jwt-audience", "", "The Google JWT audience claim for validation. If unspecified then validation is skipped")

	loggerConfig = logr.NewConfigFromFlags(cmd.Flags())
	cfg.AgentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
