package main

import (
	"context"
	"fmt"
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
		loggerConfig *logr.Config
		opts         runner.AgentOptions
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
			client, err := api.NewClient(api.Config{
				URL:           opts.URL,
				Token:         opts.Token,
				Logger:        logger,
				RetryRequests: true,
			})
			if err != nil {
				return err
			}
			opts.OperationConfig.IsAgent = true
			runner, err := runner.New(
				logger,
				&remoteClient{Client: client},
				&runner.RemoteOperationSpawner{
					Logger: logger,
					Config: opts.OperationConfig,
					URL:    opts.URL,
				},
				true,
				*opts.Config,
			)
			if err != nil {
				return fmt.Errorf("initializing agent: %w", err)
			}
			// blocks
			return runner.Start(cmd.Context())
		},
	}

	opts = runner.AgentOptions{
		Config: runner.NewConfigFromFlags(cmd.Flags()),
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Give agent a descriptive name. Optional.")
	cmd.Flags().StringVar(&opts.URL, "url", api.DefaultURL, "URL of OTF server")
	cmd.Flags().StringVar(&opts.Token, "token", "", "Agent token for authentication")

	cmd.MarkFlagRequired("token")
	cmd.SetArgs(args)

	loggerConfig = logr.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
