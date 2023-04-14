package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *CLI) stateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "State management",
	}

	cmd.AddCommand(a.stateRollbackCommand())

	return cmd
}

func (a *CLI) stateRollbackCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "rollback [id]",
		Short:         "Rollback to a previous state",
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
