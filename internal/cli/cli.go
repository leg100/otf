// Package cli provides the CLI client, i.e. the `otf` binary.
package cli

import (
	"context"
	"io"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CLI is the `otf` cli application
type CLI struct {
	client *otfhttp.Client
	creds  CredentialsStore
}

func NewCLI() *CLI {
	return &CLI{
		client: &otfhttp.Client{},
	}
}

func (a *CLI) Run(ctx context.Context, args []string, out io.Writer) error {
	var cfg otfhttp.ClientConfig

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

	cmd.PersistentFlags().StringVar(&cfg.URL, "url", otfhttp.DefaultURL, "URL of OTF server")
	cmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "API authentication token")

	cmd.SetArgs(args)
	cmd.SetOut(out)

	cmd.AddCommand(organization.NewCommand(a.client))
	cmd.AddCommand(user.NewUserCommand(a.client))
	cmd.AddCommand(user.NewTeamMembershipCommand(a.client))
	cmd.AddCommand(team.NewTeamCommand(a.client))
	cmd.AddCommand(workspace.NewCommand(a.client))
	cmd.AddCommand(run.NewCommand(a.client))
	cmd.AddCommand(state.NewCommand(a.client))
	cmd.AddCommand(runner.NewAgentsCommand(a.client))

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}

func (a *CLI) newClient(cfg *otfhttp.ClientConfig) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error {
		// Set API token according to the following precedence:
		// (1) flag
		// (2) host-specific env var
		// (3) env var
		// (4) credentials file

		if cfg.Token == "" {
			// not set via flag, so try lower precedence options
			token, err := a.getToken(cfg.URL)
			if err != nil {
				return err
			}
			cfg.Token = token
		}

		httpClient, err := otfhttp.NewClient(*cfg)
		if err != nil {
			return err
		}
		*a.client = *httpClient
		return nil
	}
}

func (a *CLI) getToken(address string) (string, error) {
	if token := os.Getenv(internal.CredentialEnvKey(address)); token != "" {
		return token, nil
	}
	if token := os.Getenv("OTF_TOKEN"); token != "" {
		return token, nil
	}
	return a.creds.Load(address)
}
