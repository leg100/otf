package cli

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func (a *CLI) stateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "State version management",
	}

	cmd.AddCommand(a.stateRollbackCommand())
	cmd.AddCommand(a.stateDownloadCommand())

	return cmd
}

func (a *CLI) stateDownloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "download [id]",
		Short:         "Download state version",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := a.DownloadState(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			var out bytes.Buffer
			if err := json.Indent(&out, state, "", "  "); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), out.String())
			return nil
		},
	}
}

func (a *CLI) stateRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "rollback [id]",
		Short:         "Rollback to a previous state version",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := a.RollbackStateVersion(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully rolled back state\n")
			return nil
		},
	}
}
