package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	defaultAddress  = ":8080"
	defaultDatabase = "postgres:///otf?host=/var/run/postgresql"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		// Stop handling ^C; another ^C will exit the program.
		cancel()
	}()

	if err := parseFlags(ctx, os.Args[1:], os.Stdout); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func parseFlags(ctx context.Context, args []string, out io.Writer) error {
	cfg := daemon.NewConfig()

	var loggerConfig logr.Config

	cmd := &cobra.Command{
		Use:           "otfd",
		Short:         "otf daemon",
		Long:          "otfd is the daemon component of the open terraforming framework.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logr.New(loggerConfig)
			if err != nil {
				return err
			}

			// Confer superuser privileges on all calls to service endpoints
			ctx := authz.AddSubjectToContext(cmd.Context(), &authz.Superuser{Username: "app-user"})

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
	cmd.Flags().BytesHexVar(&cfg.Secret, "secret", nil, "Hex-encoded 16 byte secret for cryptographic work. Required.")
	cmd.Flags().Int64Var(&cfg.MaxConfigSize, "max-config-size", cfg.MaxConfigSize, "Maximum permitted configuration size in bytes.")
	cmd.Flags().StringVar(&cfg.WebhookHost, "webhook-hostname", "", "External hostname for otf webhooks")

	// TODO: remove after given amount of time
	_ = cmd.Flags().String("allowed-origins", "", "Allowed origins for websocket upgrades")
	cmd.Flags().MarkDeprecated("allowed-origins", "websockets no longer implemented so this flag has no effect")

	cmd.Flags().StringVar(&cfg.PublicKeyPath, "public-key-path", "", "Path to public key for dynamic credentials.")
	cmd.Flags().StringVar(&cfg.PrivateKeyPath, "private-key-path", "", "Path to private key for dynamic credentials.")

	cmd.Flags().IntVar(&cfg.CacheConfig.Size, "cache-size", 0, "Maximum cache size in MB. 0 means unlimited size.")
	cmd.Flags().DurationVar(&cfg.CacheConfig.TTL, "cache-expiry", internal.DefaultCacheTTL, "Cache entry TTL.")

	cmd.Flags().DurationVar(&cfg.DeleteRunsAfter, "delete-runs-after", 0, "Delete runs older than the specified age. Specifying 0 disables run deletion.")
	cmd.Flags().DurationVar(&cfg.DeleteConfigsAfter, "delete-configs-after", 0, "Delete configs older than the specified age. Specifying 0 disables config deletion.")

	cmd.Flags().BoolVar(&cfg.SSL, "ssl", false, "Toggle SSL")
	cmd.Flags().StringVar(&cfg.CertFile, "cert-file", "", "Path to SSL certificate (required if enabling SSL)")
	cmd.Flags().StringVar(&cfg.KeyFile, "key-file", "", "Path to SSL key (required if enabling SSL)")
	cmd.Flags().BoolVar(&cfg.EnableRequestLogging, "log-http-requests", false, "Log HTTP requests")

	cmd.Flags().Var(cfg.GithubHostname, "github-hostname", "github hostname")
	cmd.Flags().StringVar(&cfg.GithubClientID, "github-client-id", "", "github client ID")
	cmd.Flags().StringVar(&cfg.GithubClientSecret, "github-client-secret", "", "github client secret")

	cmd.Flags().Var(cfg.GitlabHostname, "gitlab-hostname", "gitlab hostname")
	cmd.Flags().StringVar(&cfg.GitlabClientID, "gitlab-client-id", "", "gitlab client ID")
	cmd.Flags().StringVar(&cfg.GitlabClientSecret, "gitlab-client-secret", "", "gitlab client secret")

	cmd.Flags().Var(cfg.ForgejoHostname, "forgejo-hostname", "forgejo default hostname")

	cmd.Flags().StringVar(&cfg.OIDC.Name, "oidc-name", "", "User friendly OIDC name")
	cmd.Flags().StringVar(&cfg.OIDC.IssuerURL, "oidc-issuer-url", "", "OIDC issuer URL")
	cmd.Flags().StringVar(&cfg.OIDC.ClientID, "oidc-client-id", "", "OIDC client ID")
	cmd.Flags().StringVar(&cfg.OIDC.ClientSecret, "oidc-client-secret", "", "OIDC client secret")
	cmd.Flags().StringSliceVar(&cfg.OIDC.Scopes, "oidc-scopes", authenticator.DefaultOIDCScopes, "OIDC scopes")
	cmd.Flags().StringVar(&cfg.OIDC.UsernameClaim, "oidc-username-claim", string(authenticator.DefaultClaim), "OIDC claim to be used for username (name, email, or sub)")

	cmd.Flags().BoolVar(&cfg.RestrictOrganizationCreation, "restrict-org-creation", false, "Restrict organization creation capability to site admin role")

	cmd.Flags().StringVar(&cfg.GoogleIAPConfig.Audience, "google-jwt-audience", "", "The Google JWT audience claim for validation. If unspecified then validation is skipped")

	cmd.Flags().DurationVar(&cfg.PlanningTimeout, "planning-timeout", 2*time.Hour, "Timeout for plans.")
	cmd.Flags().DurationVar(&cfg.ApplyingTimeout, "applying-timeout", 24*time.Hour, "Timeout for applies.")

	cmd.Flags().Var(cfg.DefaultEngine, "default-engine", "Default engine for runs: terraform or tofu")

	logr.RegisterFlags(cmd.Flags(), &loggerConfig)
	runner.RegisterFlags(cmd.Flags(), cfg.RunnerConfig)

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}
