// Package cli provides the CLI client, i.e. the `otf` binary.
package cli

import (
	"context"
	"io"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CLI is the `otf` cli application
type CLI struct {
	httpClient *http.Client
	creds      CredentialsStore
}

func NewCLI() *CLI {
	return &CLI{
		httpClient: &http.Client{},
	}
}

func (a *CLI) Run(ctx context.Context, args []string, out io.Writer) error {
	cfg := http.NewConfig()

	creds, err := NewCredentialsStore()
	if err != nil {
		return err
	}
	a.creds = creds

	cmd := &cobra.Command{
		Use:               "otf",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: a.newClient(&cfg),
	}

	cmd.PersistentFlags().StringVar(&cfg.Address, "address", http.DefaultAddress, "Address of OTF server")
	cmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "API authentication token")

	cmd.SetArgs(args)
	cmd.SetOut(out)

	cmd.AddCommand(organization.NewCommand(a.httpClient))
	cmd.AddCommand(auth.NewUserCommand(a.httpClient))
	cmd.AddCommand(auth.NewTeamCommand(a.httpClient))
	cmd.AddCommand(workspace.NewCommand(a.httpClient))
	cmd.AddCommand(run.NewCommand(a.httpClient))
	cmd.AddCommand(state.NewCommand(a.httpClient))
	cmd.AddCommand(tokens.NewAgentsCommand(a.httpClient))

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}

func (a *CLI) newClient(cfg *http.Config) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		// Set API token according to the following precedence:
		// (1) flag
		// (2) host-specific env var
		// (3) env var
		// (4) credentials file

		if cfg.Token == "" {
			// not set via flag, so try lower precedence options
			token, err := a.getToken(cfg.Address)
			if err != nil {
				return err
			}
			cfg.Token = token
		}

		httpClient, err := http.NewClient(*cfg)
		if err != nil {
			return err
		}
		*a.httpClient = *httpClient
		return nil
	}
}

func (a *CLI) getToken(address string) (string, error) {
	if token, ok := os.LookupEnv(internal.CredentialEnvKey(address)); ok {
		return token, nil
	}
	if token, ok := os.LookupEnv("OTF_TOKEN"); ok {
		return token, nil
	}
	token, err := a.creds.Load(address)
	if err != nil {
		return "", err
	}
	return token, nil
}
