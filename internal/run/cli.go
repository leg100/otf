package run

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/leg100/otf/internal/configversion"
	otfhttp "github.com/leg100/otf/internal/http"

	"github.com/leg100/otf/internal/resource"

	"github.com/leg100/otf/internal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type CLI struct {
	client  cliClient
	configs cliConfigsClient
}

type cliClient interface {
	Get(ctx context.Context, runID resource.TfeID) (*Run, error)
}

type cliConfigsClient interface {
	DownloadConfig(ctx context.Context, id resource.TfeID) ([]byte, error)
}

func NewCommand(client *otfhttp.Client) *cobra.Command {
	cli := &CLI{}
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Runs management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &Client{Client: client}
			cli.configs = &configversion.Client{Client: client}
			return nil
		},
	}

	cmd.AddCommand(cli.runDownloadCommand())

	return cmd
}

func (a *CLI) runDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "download [run-id]",
		Short:         "Download configuration for run",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resource.ParseTfeID(args[0])
			if err != nil {
				return err
			}
			run, err := a.client.Get(cmd.Context(), id)
			if err != nil {
				return errors.Wrap(err, "retrieving run")
			}

			tarball, err := a.configs.DownloadConfig(cmd.Context(), run.ConfigurationVersionID)
			if err != nil {
				return errors.Wrap(err, "downloading tarball")
			}
			dest, err := os.MkdirTemp("", fmt.Sprintf("%s-*", args[0]))
			if err != nil {
				return err
			}

			reader := bytes.NewReader(tarball)
			if err := internal.Unpack(reader, dest); err != nil {
				return errors.Wrap(err, "extracting tarball")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Extracted tarball to: %s\n", dest)

			return nil
		},
	}

	return cmd
}
