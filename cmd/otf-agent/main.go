package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		// Stop handling ^C; another ^C will exit the program.
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	var (
		config *runner.Config
		url    string
		token  string
	)

	cmd := &cobra.Command{
		Use:           "otf-agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logr.New(config.LoggerConfig)
			if err != nil {
				return err
			}
			agent, err := runner.NewAgent(logger, runner.AgentOptions{
				Config: config,
				URL:    url,
				Token:  token,
			})
			if err != nil {
				return err
			}
			// blocks
			return agent.Start(cmd.Context())
		},
	}

	loggerConfig := logr.LoadConfigFromFlags(cmd.Flags())
	config = runner.LoadConfigFromFlags(cmd.Flags(), loggerConfig)
	cmd.Flags().StringVar(&config.Name, "name", "", "Give agent a descriptive name. Optional.")
	cmd.Flags().StringVar(&url, "url", api.DefaultURL, "URL of OTF server")
	cmd.Flags().StringVar(&token, "token", "", "Agent token for authentication")

	cmd.MarkFlagRequired("token")
	cmd.SetArgs(args)

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
