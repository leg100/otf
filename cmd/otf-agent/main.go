package main

import (
	"context"
	"fmt"
	"os"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	otfapi "github.com/leg100/otf/internal/api"
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
		loggerConfig *logr.Config
		clientConfig otfapi.Config
		agentConfig  *agent.Config
	)

	cmd := &cobra.Command{
		Use:           "otf-agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logr.New(loggerConfig)
			if err != nil {
				return err
			}

			// Sends unauthenticated ping to server
			app, err := agent.NewClient(clientConfig)
			if err != nil {
				return err
			}

			agent, err := agent.New(cmd.Context(), logger, app, *agentConfig)
			if err != nil {
				return fmt.Errorf("unable to start agent: %w", err)
			}
			// blocks
			return agent.Start(ctx)
		},
	}

	cmd.Flags().StringVar(&clientConfig.Address, "address", otfapi.DefaultAddress, "Address of OTF server")
	cmd.Flags().StringVar(&clientConfig.Token, "token", "", "Agent token for authentication")
	cmd.MarkFlagRequired("token")
	cmd.SetArgs(args)

	loggerConfig = logr.NewConfigFromFlags(cmd.Flags())
	agentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
