// Package cli provides the CLI client, i.e. the `otf` binary.
package cli

import (
	"context"
	"io"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/http"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CLI is the `otf` cli application
type CLI struct {
	client.Client
	creds CredentialsStore
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

	cmd.AddCommand(a.organizationCommand())
	cmd.AddCommand(a.userCommand())
	cmd.AddCommand(a.teamCommand())
	cmd.AddCommand(a.workspaceCommand())
	cmd.AddCommand(a.runCommand())
	cmd.AddCommand(a.agentCommand())
	cmd.AddCommand(a.stateCommand())

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

		client, err := client.New(*cfg)
		if err != nil {
			return err
		}
		a.Client = client
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
