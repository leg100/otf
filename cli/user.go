package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *Application) userCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "User account management",
	}

	cmd.AddCommand(a.userNewCommand())
	cmd.AddCommand(a.userDeleteCommand())

	return cmd
}

func (a *Application) userNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "new [username]",
		Short:         "Create a new user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := a.CreateUser(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created user %s\n", user.Username)
			return nil
		},
	}
}

func (a *Application) userDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "delete [username]",
		Short:         "Delete a user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.DeleteUser(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted user %s\n", args[0])
			return nil
		},
	}
}
