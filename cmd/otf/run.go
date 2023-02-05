package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/leg100/otf"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (a *application) runCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Runs management",
	}

	cmd.AddCommand(a.runDownloadCommand())

	return cmd
}

func (a *application) runDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "download [run-id]",
		Short:         "Download configuration for run",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			run, err := a.GetRun(cmd.Context(), args[0])
			if err != nil {
				return errors.Wrap(err, "retrieving run")
			}

			tarball, err := a.DownloadConfig(cmd.Context(), run.ConfigurationVersionID())
			if err != nil {
				return errors.Wrap(err, "downloading tarball")
			}
			dest, err := os.MkdirTemp("", fmt.Sprintf("%s-*", args[0]))
			if err != nil {
				return err
			}

			reader := bytes.NewReader(tarball)
			if err := otf.Unpack(reader, dest); err != nil {
				return errors.Wrap(err, "extracting tarball")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Extracted tarball to: %s\n", dest)

			return nil
		},
	}

	return cmd
}
