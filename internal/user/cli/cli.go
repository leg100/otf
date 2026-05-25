package cli

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/user"
	userapi "github.com/leg100/otf/internal/user/api"
	"github.com/spf13/cobra"
)

type userCLI struct {
	client userCLIClient
}

type userCLIClient interface {
	Create(ctx context.Context, username string, opts ...user.NewUserOption) (*user.User, error)
	Delete(ctx context.Context, username user.Username) error
}

func NewUserCommand(apiClient *otfhttp.Client) *cobra.Command {
	cli := &userCLI{}
	cmd := &cobra.Command{
		Use:   "users",
		Short: "User account management",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return err
			}
			cli.client = &userapi.Client{Client: apiClient}
			return nil
		},
	}

	cmd.AddCommand(cli.userNewCommand())
	cmd.AddCommand(cli.userDeleteCommand())

	return cmd
}

func (a *userCLI) userNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "new [username]",
		Short:         "Create a new user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := a.client.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully created user %s\n", user.Username)
			return nil
		},
	}
}

func (a *userCLI) userDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "delete [username]",
		Short:         "Delete a user account",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := user.NewUsername(args[0])
			if err != nil {
				return err
			}
			if err := a.client.Delete(cmd.Context(), username); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Successfully deleted user %s\n", args[0])
			return nil
		},
	}
}
