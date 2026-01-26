package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/runner/agent"
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
		config       = runner.NewDefaultConfig()
		loggerConfig logr.Config
		url          string
		token        string
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
			// if using the kubernetes executor then the server url should be
			// set to the value of the --url flag
			if config.ExecutorKind == runner.KubeExecutorKind {
				config.KubeConfig.ServerURL = url
			}
			// Construct runner.
			runner, err := agent.New(logger, url, token, config)
			if err != nil {
				return err
			}
			// blocks
			return runner.Start(cmd.Context())
		},
	}

	logr.RegisterFlags(cmd.Flags(), &loggerConfig)
	runner.RegisterFlags(cmd.Flags(), config)

	cmd.Flags().StringVar(&config.Name, "name", "", "Give agent a descriptive name. Optional.")
	cmd.Flags().StringVar(&url, "url", otfhttp.DefaultURL, "URL of OTF server")
	cmd.Flags().StringVar(&token, "token", "", "Agent token for authentication")

	cmd.MarkFlagRequired("token")
	cmd.SetArgs(args)

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
