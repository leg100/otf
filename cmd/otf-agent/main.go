package main

import (
	"context"
	"fmt"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := context.WithCancel(context.Background())
	cmdutil.CatchCtrlC(cancel)

	if err := run(ctx, os.Args[1:]); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	var (
		loggerCfg *logr.Config
		cfg       *agent.Config
	)

	clientCfg := http.NewConfig()

	cmd := &cobra.Command{
		Use:           "otf-agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logr.New(loggerCfg)
			if err != nil {
				return err
			}

			// Sends unauthenticated ping to server
			app, err := client.New(*clientCfg)
			if err != nil {
				return err
			}

			// Confirm token validity
			at, err := app.GetAgentToken(ctx, "")
			if err != nil {
				return fmt.Errorf("attempted authentication: %w", err)
			}
			logger.Info("successfully authenticated", "organization", at.Organization, "token_id", at.ID)

			// Ensure agent only processes runs for this org
			cfg.Organization = internal.String(at.Organization)

			agent, err := agent.NewAgent(logger, app, *cfg)
			if err != nil {
				return fmt.Errorf("unable to start agent: %w", err)
			}
			// blocks
			return agent.Start(ctx)
		},
	}
	cmd.Flags().StringVar(&clientCfg.Address, "address", http.DefaultAddress, "Address of OTF server")
	cmd.Flags().StringVar(&clientCfg.Token, "token", "", "Agent token for authentication")
	cmd.MarkFlagRequired("token")

	cmd.SetArgs(args)

	loggerCfg = logr.NewConfigFromFlags(cmd.Flags())
	cfg = agent.NewConfigFromFlags(cmd.Flags())
	// otf-agent is an 'external' agent, as opposed to the internal agent in
	// otfd.
	cfg.External = true

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
