package main

import (
	"context"
	"io"

	"github.com/leg100/otf/client"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/http"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type application struct {
	client.Client
}

func (a *application) run(ctx context.Context, args []string, out io.Writer) error {
	cfg, err := http.NewConfig(LoadCredentials)
	if err != nil {
		return err
	}

	cmd := &cobra.Command{
		Use:           "otf",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(*cfg)
			if err != nil {
				return err
			}
			a.Client = client
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.Address, "address", http.DefaultAddress, "Address of OTF server")

	cmd.SetArgs(args)
	cmd.SetOut(out)

	cmd.AddCommand(a.organizationCommand())
	cmd.AddCommand(a.workspaceCommand())
	cmd.AddCommand(a.runCommand())
	cmd.AddCommand(a.agentCommand())

	if err = cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
