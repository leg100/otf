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
	"github.com/leg100/otf/internal/resource"
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
		opts         *runner.AgentOptions
		jobToken     string
		jobIDRaw     string
	)

	cmd := &cobra.Command{
		Use:           "otf-job",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID, err := resource.ParseTfeID(jobIDRaw)
			if err != nil {
				return err
			}
			logger, err := logr.New(loggerConfig)
			if err != nil {
				return err
			}
			spawner := &runner.RemoteOperationSpawner{
				Config: opts.OperationConfig,
				Logger: logger,
				URL:    opts.URL,
			}
			// blocks until job completion
			_, err = spawner.NewOperation(cmd.Context(), jobID, []byte(jobToken))
			return err
		},
	}

	loggerConfig = logr.NewConfigFromFlags(cmd.Flags())
	cmd.Flags().StringVar(&jobToken, "job-token", "", "Job token for authentication")
	cmd.Flags().StringVar(&jobIDRaw, "job-id", "", "ID of job to execute")
	cmd.Flags().StringVar(&opts.URL, "url", api.DefaultURL, "URL of OTF server")

	cmd.SetArgs(args)

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
