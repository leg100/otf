// Package cli provides the CLI client, i.e. the `otf` binary.
package cli

import (
	"context"
	"io"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/remoteops"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CLI is the `otf` cli application
type CLI struct {
	api   *api.Client
	creds CredentialsStore
}

func NewCLI() *CLI {
	return &CLI{
		api: &api.Client{},
	}
}

func (a *CLI) Run(ctx context.Context, args []string, out io.Writer) error {
	var cfg api.Config

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

	cmd.PersistentFlags().StringVar(&cfg.Address, "address", api.DefaultAddress, "Address of OTF server")
	cmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "API authentication token")

	cmd.SetArgs(args)
	cmd.SetOut(out)

	cmd.AddCommand(organization.NewCommand(a.api))
	cmd.AddCommand(user.NewUserCommand(a.api))
	cmd.AddCommand(user.NewTeamMembershipCommand(a.api))
	cmd.AddCommand(team.NewTeamCommand(a.api))
	cmd.AddCommand(workspace.NewCommand(a.api))
	cmd.AddCommand(run.NewCommand(a.api))
	cmd.AddCommand(state.NewCommand(a.api))
	cmd.AddCommand(remoteops.NewAgentsCommand(a.api))

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}

func (a *CLI) newClient(cfg *api.Config) func(*cobra.Command, []string) error {
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

		httpClient, err := api.NewClient(*cfg)
		if err != nil {
			return err
		}
		*a.api = *httpClient
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
